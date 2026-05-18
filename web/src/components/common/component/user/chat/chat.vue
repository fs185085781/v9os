<script setup>
import { defineAsyncComponent, onBeforeUnmount, onMounted, ref } from "vue";
import { NButton, NEmpty, NScrollbar, NTag } from "naive-ui";
import { postData, uuid } from "@/util/util";

let EnterpriseChat = null;
const mod = window.__INFACE_MODS__?.["/src/components/common/inface/user_ee/chat.vue"];
if (mod) {
  EnterpriseChat = defineAsyncComponent(mod);
}

const props = defineProps({
  data: {},
  winId: {
    type: String,
    default: "",
  },
});

const notices = ref([]);
const wsListenerId = `base-chat-${uuid()}`;

async function loadNotices() {
  const res = await postData("chat", "notices", { page: 1, pageSize: 200 }, "");
  notices.value = Array.isArray(res) ? res : [];
}

function parseMessage(raw) {
  let data = raw;
  if (typeof data === "string") {
    try {
      data = JSON.parse(data);
    } catch {
      return null;
    }
  }
  return data;
}

function onWsMessage(raw) {
  const msg = parseMessage(raw);
  if (!msg || msg.type === "chat_message") return;
  notices.value.unshift({
    ID: `local-${Date.now()}`,
    Type: msg.type || "notice",
    Msg: msg.msg || "",
    DateTime: msg.date_time || Date.now(),
  });
}

function formatTime(time) {
  if (!time) return "";
  return new Date(Number(time)).toLocaleString();
}

async function clearNotice(item) {
  if (String(item.ID).startsWith("local-")) {
    notices.value = notices.value.filter((row) => row.ID !== item.ID);
    return;
  }
  const ok = await postData("chat", "delete_notices", { ids: [item.ID] }, "okerr");
  if (ok) loadNotices();
}

onMounted(() => {
  if (!EnterpriseChat) {
    loadNotices();
    window.$websocket.addClient("/chat", wsListenerId, null, onWsMessage, null);
  }
});

onBeforeUnmount(() => {
  if (!EnterpriseChat) {
    window.$websocket.removeClient("/chat", wsListenerId);
  }
});
</script>

<template>
  <EnterpriseChat v-if="EnterpriseChat" :data="props.data || {}" :win-id="props.winId" />
  <div v-else class="h-full min-h-0 flex flex-col user-color-fbg user-color-ftext">
    <header class="h-16 flex items-center justify-between border-b user-color-line px-4.5">
      <div>
        <div class="text-16px font-600">系统通知</div>
        <div class="mt-0.5 text-12px user-color-muted">无企业模块时聊天中心仅显示系统通知</div>
      </div>
      <n-button size="small" @click="loadNotices">刷新</n-button>
    </header>
    <n-scrollbar class="min-h-0 flex-1 p-3">
      <div v-for="item in notices" :key="item.ID" class="mb-2.5 flex items-start gap-3 user-rounded-1.5 border user-color-border user-color-readable p-3">
        <div class="min-w-0 flex-1">
          <div class="mb-1.5 flex items-center gap-2 text-12px user-color-muted">
            <n-tag size="small" type="info">{{ item.Type || item.type || "notice" }}</n-tag>
            <span>{{ formatTime(item.DateTime || item.date_time) }}</span>
          </div>
          <div class="whitespace-pre-wrap break-words leading-1.55">{{ item.Msg || item.msg }}</div>
        </div>
        <n-button size="small" quaternary @click="clearNotice(item)">删除</n-button>
      </div>
      <n-empty v-if="notices.length === 0" description="暂无系统通知" class="mt-20" />
    </n-scrollbar>
  </div>
</template>
