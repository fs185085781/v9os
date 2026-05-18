<script setup>
import { onMounted, onUnmounted, ref, computed } from "vue";
import { getWinSize } from "@/util/util.js";
import { useEventBus } from "@/util/event.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/IconView.vue";

const myApps = ref([]);

onMounted(async () => {
  myApps.value = await $user.getDesktopApps();
});
useEventBus("desktop-apps-change", async () => {
  myApps.value = await $user.getDesktopApps();
});

const visibleApps = computed(() => {
  return myApps.value;
});

const appName = (app) => {
  return app.name?.value || app.name || app.code;
};

const openApp = (app) => {
  const sz = getWinSize();
  const options = {
    id: app.windowId || "",
    width: sz.width,
    height: sz.height,
    title: appName(app),
    icon: app.icon,
    data: {},
  };
  if (app.type === "system") {
    options.component = app.url;
  } else if (app.type === "plugin" || app.type === "url") {
    options.iframeUrl = app.url;
  }
  $wins.addWindow(options);
};

async function deleteShortcut(app) {
  if (!app?.id) return;
  if (await $msg.util.confirm("确定删除该桌面快捷方式吗？")) {
    await $user.deleteDesktopApp(app.id);
  }
}

const appContext = (app) => ({
  type: "win10.desktop.shortcut",
  payload: app,
  actions: {
    edit: () => openDesktopShortcutWin(app, { title: "编辑快捷方式" }),
    delete: () => deleteShortcut(app),
  },
});
const appMenuResolver = () => [
  {
    key: "edit",
    label: () => "编辑快捷方式",
    group: "main",
    order: 10,
    action: (ctx) => ctx.actions.edit(),
  },
  {
    key: "delete",
    label: () => "删除快捷方式",
    group: "main",
    order: 20,
    action: (ctx) => ctx.actions.delete(),
  },
];
$contextMenu.register("win10.desktop.shortcut", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("win10.desktop.shortcut", appMenuResolver);
});
</script>

<template>
  <div class="win10-desktop-icons fixed left-0 top-0 bottom-10 z-1 p-3 flex flex-col flex-wrap content-start gap-y-3">
    <button
      v-for="app in visibleApps"
      :key="app.code"
      class="win10-desktop-icon"
      @dblclick="openApp(app)"
      @click.stop
      v-context-menu="appContext(app)"
    >
      <IconView :icon="app.icon" :alt="appName(app)" :size="36" icon-class="w-9 h-9 object-contain mx-auto" />
      <span class="mt-1 text-xs text-white leading-4 break-words line-clamp-2">{{ appName(app) }}</span>
    </button>
  </div>
</template>
