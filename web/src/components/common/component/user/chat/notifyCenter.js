import { computed, defineAsyncComponent, reactive } from "vue";
import { getBigData, setBigData } from "@/util/bigdata";
import { getWinSize, postData, uuid } from "@/util/util";
import emitter from "@/util/event";
import { MessageOutlined } from "@vicons/antd";
import { renderIcon } from "@/util/icon";

const dayMs = 24 * 60 * 60 * 1000;
const enterpriseChatPath = "/src/components/common/inface/user_ee/chat.vue";
const enterpriseMessagePath = "/src/components/common/inface/user_ee/wins/chat/message.js";

function dateKey(time) {
  const d = new Date(time);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}${m}${day}`;
}

let bucketRange = function fallbackBucketRange(time) {
  const d = new Date(Number(time || Date.now()));
  d.setHours(0, 0, 0, 0);
  const day = d.getDay() || 7;
  const start = d.getTime() - (day - 1) * dayMs;
  return { start, end: start + 6 * dayMs, startKey: dateKey(start), endKey: dateKey(start + 6 * dayMs) };
};

let normalizeId = function fallbackNormalizeId(item) {
  return Number(item.ID || item.id || item.createdAt || item.CreatedAt || Date.now());
};

let normalizeTime = function fallbackNormalizeTime(time) {
  if (!time) return Date.now();
  if (typeof time === "number") return time;
  const num = Number(time);
  if (!Number.isNaN(num) && num > 0) return num;
  const parsed = Date.parse(time);
  return Number.isNaN(parsed) ? Date.now() : parsed;
};

let messageDay = function fallbackMessageDay(item) {
  const time = normalizeTime(item.createdAt || item.CreatedAt || Date.now());
  return new Date(time).toISOString().slice(0, 10);
};

let normalizeMessage = function fallbackNormalizeMessage(item) {
  return {
    id: item.ID || item.id || `${item.fromUserId || item.FromUserID}-${Date.now()}`,
    fromUserId: Number(item.FromUserID || item.fromUserId || item.from || 0),
    toUserId: Number(item.ToUserID || item.toUserId || item.to || 0),
    groupId: Number(item.GroupID || item.groupId || 0),
    msgType: item.MsgType || item.msgType || item.type || "text",
    content: item.Content || item.content || item.msg || "",
    createdAt: normalizeTime(item.CreatedAt || item.createdAt || item.date_time || item.DateTime || Date.now()),
    readAt: Number(item.ReadAt || item.readAt || 0),
  };
};

let messageContent = function fallbackMessageContent(item) {
  if (item.msgType === "notice") return item.content;
  try {
    const obj = JSON.parse(item.content || "{}");
    if (item.msgType === "text") return obj.text || "";
    if (item.msgType === "image") return `[图片] ${obj.name || obj.url || ""}`;
    if (item.msgType === "file") return `[文件] ${obj.name || ""}`;
    if (item.msgType === "url") return `[链接] ${obj.title || obj.url || ""}`;
    if (item.msgType === "card") return `[卡片] ${obj.title || ""}`;
  } catch {
    return item.content || "";
  }
  return item.content || "";
};

let enterpriseUtilsLoaded = false;

function hasEnterpriseChat() {
  const mods = window.__INFACE_MODS__ || {};
  return !!(mods[enterpriseChatPath] || mods[enterpriseMessagePath]);
}

async function loadEnterpriseMessageUtils() {
  if (enterpriseUtilsLoaded) return;
  enterpriseUtilsLoaded = true;
  const modLoader = window.__INFACE_MODS__?.[enterpriseMessagePath];
  if (!modLoader) return;
  const mod = await modLoader();
  bucketRange = mod.bucketRange || bucketRange;
  messageContent = mod.messageContent || messageContent;
  messageDay = mod.messageDay || messageDay;
  normalizeId = mod.normalizeId || normalizeId;
  normalizeMessage = mod.normalizeMessage || normalizeMessage;
  normalizeTime = mod.normalizeTime || normalizeTime;
}

const state = reactive({
  inited: false,
  loading: false,
  sendFn: null,
  messages: [],
  notices: [],
  bucketKeys: [],
});

const listenerId = `topbar-chat-${uuid()}`;
const chatWindowId = "v9os-chat-win";

function currentUserId() {
  return Number(window.$user?.user?.ID || globalThis.$user?.user?.ID || 0);
}

function chatIndexKey() {
  return `chat-ee-index-${currentUserId()}`;
}

function chatBucketKeyByRange(range) {
  return `chat-ee-${currentUserId()}-${range.startKey}-${range.endKey}`;
}

function chatBucketKeyByMessage(item) {
  return chatBucketKeyByRange(bucketRange(item.createdAt));
}

async function loadChatIndex() {
  const rows = await getBigData(chatIndexKey());
  state.bucketKeys = Array.isArray(rows) ? rows.filter(Boolean).sort() : [];
}

async function saveChatIndex() {
  await setBigData(chatIndexKey(), [...new Set(state.bucketKeys)].sort());
}

async function readBucket(key) {
  const rows = await getBigData(key);
  return Array.isArray(rows) ? rows.map(normalizeMessage) : [];
}

async function writeBucket(key, rows) {
  await setBigData(key, rows.sort((a, b) => normalizeId(a) - normalizeId(b)));
}

async function mergeMessages(rows) {
  const grouped = {};
  rows.forEach((raw) => {
    const item = normalizeMessage(raw);
    if (!item.content) return;
    const key = chatBucketKeyByMessage(item);
    if (!grouped[key]) grouped[key] = [];
    grouped[key].push(item);
  });
  const changedKeys = Object.keys(grouped);
  for (const key of changedKeys) {
    const oldRows = await readBucket(key);
    const map = {};
    oldRows.forEach((item) => {
      map[String(item.id)] = item;
    });
    grouped[key].forEach((item) => {
      const old = map[String(item.id)];
      const next = { ...old, ...item };
      if (Number(old?.readAt || 0) > Number(next.readAt || 0)) next.readAt = old.readAt;
      map[String(item.id)] = next;
    });
    await writeBucket(key, Object.values(map));
  }
  if (changedKeys.length > 0) {
    state.bucketKeys = [...new Set([...state.bucketKeys, ...changedKeys])].sort();
    await saveChatIndex();
  }
}

async function loadLocalMessages() {
  const rows = [];
  for (const key of state.bucketKeys.slice(-12)) {
    rows.push(...await readBucket(key));
  }
  state.messages = rows.sort((a, b) => normalizeId(b) - normalizeId(a));
}

async function syncServerMessages() {
  if (!hasEnterpriseChat()) return;
  let page = 1;
  const pageSize = 1000;
  for (;;) {
    const res = await postData("chat_ee", "messages", { all: true, page, pageSize }, "");
    if (!Array.isArray(res) || res.length === 0) break;
    await mergeMessages(res);
    if (res.length < pageSize) break;
    page++;
  }
}

async function loadNotices() {
  const res = await postData("chat", "notices", { page: 1, pageSize: 200 }, "");
  state.notices = Array.isArray(res) ? res : [];
}

function parseWsMessage(raw) {
  let data = raw;
  if (typeof data === "string") {
    try {
      data = JSON.parse(data);
    } catch {
      return null;
    }
  }
  if (data?.type === "chat_message" && data.msg) {
    try {
      return normalizeMessage(JSON.parse(data.msg));
    } catch {
      return null;
    }
  }
  if (data?.type === "notice" && data.msg) {
    state.notices.unshift({
      ID: `local-${Date.now()}`,
      Type: data.type,
      Msg: data.msg,
      DateTime: data.date_time || Date.now(),
    });
  }
  return null;
}

async function onWsMessage(raw) {
  await loadEnterpriseMessageUtils();
  const enterpriseChat = hasEnterpriseChat();
  const msg = parseWsMessage(raw);
  if (!enterpriseChat) return;
  if (!msg) return;
  await mergeMessages([msg]);
  await loadLocalMessages();
}

async function refresh() {
  if (!currentUserId()) return;
  state.loading = true;
  await loadEnterpriseMessageUtils();
  if (hasEnterpriseChat()) {
    await loadChatIndex();
    await syncServerMessages();
    await loadLocalMessages();
  } else {
    state.messages = [];
    state.bucketKeys = [];
  }
  await loadNotices();
  state.loading = false;
}

function initChatNotifyCenter() {
  if (state.inited) return state;
  if (!window.$websocket) {
    setTimeout(initChatNotifyCenter, 500);
    return state;
  }
  state.inited = true;
  window.$websocket.addClient("/chat", listenerId, (fn) => {
    state.sendFn = fn;
  }, onWsMessage, () => {
    state.sendFn = null;
  });
  refresh();
  emitter.on("chat-read-over", () => {
    refresh();
    setTimeout(refresh, 150);
  });
  return state;
}

function unreadMessages() {
  const uid = currentUserId();
  return state.messages.filter((item) => (item.toUserId === uid || item.groupId > 0) && item.fromUserId !== uid && Number(item.readAt || 0) <= 0);
}

function unreadRowsForPanel() {
  const chatRows = unreadMessages().map((item) => ({
    id: `chat-${item.id}`,
    kind: "chat",
    title: item.groupId > 0 ? `群聊 ${item.groupId}` : `用户 ${item.fromUserId}`,
    subtitle: messageContent(item),
    time: item.createdAt,
    unread: true,
    raw: item,
  }));
  const noticeRows = state.notices.map((item) => ({
    id: `notice-${item.ID}`,
    kind: "notice",
    title: "系统通知",
    subtitle: item.Msg || item.msg || "",
    time: item.DateTime || item.date_time || Date.now(),
    unread: true,
    raw: item,
  }));
  return [...chatRows, ...noticeRows].sort((a, b) => normalizeTime(b.time) - normalizeTime(a.time)).slice(0, 80);
}

function openChatFromMessage(row) {
  const sz = getWinSize();
  const raw = row.raw || {};
  const target = row.kind === "notice"
    ? { type: "system" }
    : {
        type: raw.groupId > 0 ? "group" : "user",
        id: raw.groupId > 0 ? raw.groupId : raw.fromUserId,
        messageId: raw.id,
      };
  $wins.addWindow({
    id: chatWindowId,
    width: sz.width,
    height: sz.height,
    title: "聊天",
    icon: renderIcon(MessageOutlined, 40),
    component: defineAsyncComponent(() => import("@/components/common/component/user/chat/chat.vue")),
    data: { target },
  });
}

function openChatCenter() {
  const sz = getWinSize();
  $wins.addWindow({
    id: chatWindowId,
    width: sz.width,
    height: sz.height,
    title: "聊天",
    icon: renderIcon(MessageOutlined, 40),
    component: defineAsyncComponent(() => import("@/components/common/component/user/chat/chat.vue")),
    data: {},
  });
}

export function useChatNotifyCenter() {
  initChatNotifyCenter();
  return {
    state,
    unreadCount: computed(() => unreadRowsForPanel().length),
    rows: computed(unreadRowsForPanel),
    refresh,
    openChatCenter,
    openChatFromMessage,
    formatDay: messageDay,
  };
}
