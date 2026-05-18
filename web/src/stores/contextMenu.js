import { reactive } from "vue";
import { defineStore } from "pinia";
import { getContextByElement } from "@/directives/contextMenu.js";

const groupOrder = {
  main: 10,
  clipboard: 20,
  create: 30,
  window: 40,
  manage: 50,
  danger: 90,
};

const callMaybe = (value, ...args) => {
  if (typeof value === "function") {
    return value(...args);
  }
  return value;
};

const mimeMatch = (mime, accept) => {
  if (!mime || !accept) {
    return false;
  }
  if (mime === accept) {
    return true;
  }
  if (accept.endsWith("/*")) {
    return mime.startsWith(accept.slice(0, -1));
  }
  return false;
};

const jsonClone = (data) => {
  data = JSON.parse(JSON.stringify(data));
  return data
};

export const contextMenuStore = defineStore("contextMenu", () => {
  const state = reactive({
    show: false,
    x: 0,
    y: 0,
    context: null,
    items: [],
    clipboard: {
      source: "",
      items: [],
    },
  });
  const resolvers = new Map();
  const contributions = new Map();
  let inited = false;

  const register = (type, resolver) => {
    if (!resolvers.has(type)) {
      resolvers.set(type, []);
    }
    resolvers.get(type).push(resolver);
  };

  const unregister = (type, resolver) => {
    if (!resolvers.has(type)) {
      return;
    }
    resolvers.set(
      type,
      resolvers.get(type).filter((item) => item !== resolver),
    );
  };

  const contribute = (slot, resolver) => {
    if (!contributions.has(slot)) {
      contributions.set(slot, []);
    }
    contributions.get(slot).push(resolver);
  };

  const uncontribute = (slot, resolver) => {
    if (!contributions.has(slot)) {
      return;
    }
    contributions.set(
      slot,
      contributions.get(slot).filter((item) => item !== resolver),
    );
  };

  const matchClipboard = (accepts = []) => {
    const list = Array.isArray(accepts) ? accepts : [accepts];
    return state.clipboard.items.filter((item) =>
      list.some((accept) => mimeMatch(item.mime, accept)),
    );
  };

  const bestClipboard = (accepts = []) => {
    const matched = matchClipboard(accepts);
    return matched.length > 0 ? matched[0] : null;
  };

  const hasCapability = (ctx, capability) => {
    if (!capability) {
      return true;
    }
    return (ctx.capabilities || []).includes(capability);
  };

  const canAccept = (accepts) => {
    if (!accepts || accepts.length === 0) {
      return true;
    }
    return matchClipboard(accepts).length > 0;
  };

  const sortItems = (items) => {
    return items.sort((a, b) => {
      const groupA = groupOrder[a.group || "main"] || 50;
      const groupB = groupOrder[b.group || "main"] || 50;
      if (groupA !== groupB) {
        return groupA - groupB;
      }
      const orderA = a.order ?? 100;
      const orderB = b.order ?? 100;
      if (orderA !== orderB) {
        return orderA - orderB;
      }
      return String(a.key || "").localeCompare(String(b.key || ""));
    });
  };

  const normalizeItems = (items, ctx) => {
    const result = [];
    for (const item of items || []) {
      if (!item) {
        continue;
      }
      if (item.type === "separator") {
        result.push(item);
        continue;
      }
      if (item.capability && !hasCapability(ctx, item.capability)) {
        continue;
      }
      if (item.accept && !canAccept(item.accept)) {
        continue;
      }
      if (callMaybe(item.visible, ctx) === false) {
        continue;
      }
      let children = [];
      if (item.childrenSlot && contributions.has(item.childrenSlot)) {
        for (const resolver of contributions.get(item.childrenSlot)) {
          children = children.concat(callMaybe(resolver, ctx) || []);
        }
      }
      if (item.children) {
        children = children.concat(callMaybe(item.children, ctx) || []);
      }
      children = normalizeItems(children, ctx);
      if ((item.children || item.childrenSlot) && children.length === 0) {
        continue;
      }
      result.push({
        ...item,
        label: callMaybe(item.label, ctx) || "",
        disabled: callMaybe(item.disabled, ctx) === true,
        children,
      });
    }
    return sortItems(result);
  };

  const buildItems = (ctx) => {
    let items = [];
    if (resolvers.has(ctx.type)) {
      for (const resolver of resolvers.get(ctx.type)) {
        items = items.concat(callMaybe(resolver, ctx) || []);
      }
    }
    if (ctx.items) {
      items = items.concat(callMaybe(ctx.items, ctx) || []);
    }
    return normalizeItems(items, ctx);
  };

  const calcMenuHeight = (items = []) => {
    const itemHeight = 32;
    const separatorHeight = 12;
    const padding = 12;
    let height = padding;
    for (const item of items) {
      if (item?.type === "separator") {
        height += separatorHeight;
      } else {
        height += itemHeight;
      }
    }
    return Math.max(48, height);
  };

  const fixPosition = (x, y, items = []) => {
    const width = 240;
    const height = calcMenuHeight(items);
    const offsetY = y - 45;
    return {
      x: Math.min(x, window.innerWidth - width - 8),
      y: Math.max(8, Math.min(offsetY, window.innerHeight - height - 8)),
    };
  };

  const show = (ctx, eventOrPoint) => {
    if (!ctx || !ctx.type) {
      return;
    }
    const event = eventOrPoint || {};
    const context = {
      payload: {},
      accepts: [],
      capabilities: [],
      actions: {},
      ...ctx,
      target: ctx.target || event.target,
      event,
      hasCapability(capability) {
        return hasCapability(this, capability);
      },
      acceptsClipboard(accepts) {
        return matchClipboard(accepts).length > 0;
      },
    };
    const items = buildItems(context);
    if (items.length === 0) {
      close();
      return;
    }
    const point = fixPosition(
      event.clientX ?? event.x ?? 0,
      event.clientY ?? event.y ?? 0,
      items,
    );
    state.context = context;
    state.items = items;
    state.x = point.x;
    state.y = point.y;
    state.show = true;
  };

  const close = () => {
    state.show = false;
    state.context = null;
    state.items = [];
  };

  const run = async (item) => {
    if (!item || item.disabled || item.type === "separator") {
      return;
    }
    if (item.children && item.children.length > 0) {
      return;
    }
    const ctx = state.context;
    const clipboardItem = item.accept ? bestClipboard(item.accept) : null;
    close();
    if (typeof item.action === "function") {
      return await item.action(ctx, {
        args: item.args || {},
        clipboardItem,
      });
    }
    if (item.actionId && ctx?.pluginWinId) {
      const iframeRef = window.__v9osIframes?.[ctx.pluginWinId];
      const iframe = iframeRef?.value?.contentWindow;
      if (iframe) {
        iframe.postMessage(
          {
            action: "context-menu-action",
            data: {
              actionId: item.actionId,
              payload: jsonClone(ctx.payload),
              args: jsonClone(item.args || {}),
              clipboardItem: jsonClone(clipboardItem),
            },
          },
          "*",
        );
      }
    }
  };

  const normalizePluginItems = (items = []) => {
    return items.map((item) => {
      if (!item) {
        return item;
      }
      const next = { ...item };
      if (typeof next.action === "string" && !next.actionId) {
        next.actionId = next.action;
        delete next.action;
      }
      if (next.children) {
        next.children = normalizePluginItems(next.children);
      }
      return next;
    });
  };

  const showFromPlugin = (winId, options = {}) => {
    const iframeRef = window.__v9osIframes?.[winId];
    const rect = iframeRef?.value?.getBoundingClientRect?.();
    show(
      {
        type: options.type || "plugin",
        scope: "plugin",
        pluginWinId: winId,
        payload: options.payload || {},
        accepts: options.accepts || [],
        capabilities: options.capabilities || [],
        items: normalizePluginItems(options.items || []),
      },
      {
        clientX: (rect?.left || 0) + (options.x || 0),
        clientY: (rect?.top || 0) + (options.y || 0),
        target: iframeRef?.value || null,
      },
    );
    return true;
  };

  const setClipboard = (clipboard) => {
    state.clipboard = {
      source: clipboard?.source || "",
      items: clipboard?.items || [],
    };
  };

  const getClipboard = () => {
    return state.clipboard;
  };

  const init = () => {
    if (inited) {
      return;
    }
    inited = true;
    document.addEventListener("contextmenu", (event) => {
      const ctx = getContextByElement(event.target);
      if (!ctx) {
        close();
        return;
      }
      event.preventDefault();
      show(ctx, event);
    });
    document.addEventListener("mousedown", () => {
      close();
    });
  };

  const api = {
    state,
    register,
    unregister,
    contribute,
    uncontribute,
    show,
    close,
    run,
    setClipboard,
    getClipboard,
    matchClipboard,
    bestClipboard,
    showFromPlugin,
    init,
  };
  window.$contextMenu = api;
  return api;
});
