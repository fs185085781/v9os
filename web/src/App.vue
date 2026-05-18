<script setup>
import { defineAsyncComponent, ref, onMounted } from "vue";
import { useStore } from "@/stores/user.js";
import { NConfigProvider } from "naive-ui";
const Macos = defineAsyncComponent(() => import("./components/macos/App.vue"));
const Win10 = defineAsyncComponent(() => import("./components/win10/App.vue"));
const Deepin = defineAsyncComponent(() => import("./components/deepin/App.vue"));
const Pad = defineAsyncComponent(() => import("./components/pad/App.vue"));
const Backend = defineAsyncComponent(
  () => import("./components/backend/App.vue"),
);
const Expand = defineAsyncComponent(
  () => import("./components/common/Expand.vue"),
);
const user = useStore();
import "@unocss/reset/tailwind.css";
import "uno.css";
import "./main.css";
import emitter from "@/util/event.js";
import { delayAction } from "@/util/util.js";
document.addEventListener("mousedown", (event) => {
  emitter.emit("document-click", event.target);
});
import { websocketStore } from "@/stores/websocket.js";
import { webhookStore } from "@/stores/webhook.js";
import { windowsStore } from "@/stores/windows.js";
import { contextMenuStore } from "@/stores/contextMenu.js";
windowsStore();
contextMenuStore().init();
websocketStore();
webhookStore().refresh();
const expandUrl = ref("");
const initExpand = () => {
  const url = new URL(window.location.href);
  if (url.searchParams.get("expand") == "true") {
    expandUrl.value = url.searchParams.get("url");
  }
};
initExpand();
const KernelUpgradeIndex = ref(null);
onMounted(() => {
  delayAction(() => {
    return $user.user && $user.user.ID > 0
  },() => {
    const mod = window.__INFACE_MODS__?.["/src/components/common/inface/kernel_update_ee/index.vue"];
    if (mod && $user.user && $user.user.ID > 0 && $user.user.IsAdmin == 1) {
      KernelUpgradeIndex.value = defineAsyncComponent(mod);
    }
  })
});
</script>
<template>
  <n-config-provider :theme-overrides="user.settings.ThemeOverride" :theme="user.settings.NaiveTheme"
    :locale="user.settings.Lang1" :date-locale="user.settings.Lang2" class="w-full h-full overflow-hidden">
    <Expand v-if="expandUrl" :url="expandUrl" />
    <Backend v-if="!expandUrl && user.settings.Mode === 'backend'" />
    <div class="v9os-desktop-root w-full h-full overflow-hidden">
      <Macos v-if="!expandUrl && user.settings.Mode === 'macos'" />
      <Win10 v-if="!expandUrl && user.settings.Mode === 'win10'" />
      <Deepin v-if="!expandUrl && user.settings.Mode === 'deepin'" />
    </div>
    <Pad v-if="!expandUrl && user.settings.Mode === 'pad'" />
    <template v-if="KernelUpgradeIndex">
      <KernelUpgradeIndex />
    </template>
  </n-config-provider>
</template>
