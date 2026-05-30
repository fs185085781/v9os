<script setup>
import { onMounted, onUnmounted, ref, computed } from "vue";
import { getWinSize } from "@/util/util.js";
import { useEventBus } from "@/util/event.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/component/util/IconView.vue";

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
  type: "deepin.desktop.shortcut",
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
$contextMenu.register("deepin.desktop.shortcut", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("deepin.desktop.shortcut", appMenuResolver);
});
</script>

<template>
  <div class="fixed left-0 top-0 bottom-20 z-1 p-3 flex flex-col flex-wrap content-start gap-y-3 max-h-[calc(100vh-5rem)]">
    <button
      v-for="app in visibleApps"
      :key="app.code"
      class="w-20 h-22 flex flex-col items-center justify-center p-1 text-center text-white bg-transparent cursor-default user-rounded-2 [text-shadow:0_1px_5px_rgba(0,0,0,0.75)] hover:bg-white/18 hover:outline hover:outline-1 hover:outline-white/22 focus:bg-white/18 focus:outline focus:outline-1 focus:outline-white/22"
      @dblclick="openApp(app)"
      @click.stop
      v-context-menu="appContext(app)"
    >
      <IconView :icon="app.icon" :alt="appName(app)" :size="44" icon-class="w-11 h-11 object-contain mx-auto" />
      <span class="mt-1 text-xs text-white leading-4 break-words line-clamp-2">{{ appName(app) }}</span>
    </button>
  </div>
</template>
