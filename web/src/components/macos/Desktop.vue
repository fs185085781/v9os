<script setup>
import { defineAsyncComponent, computed, onMounted, onUnmounted, ref } from "vue";
import TopBar from "./TopBar.vue";
import Dock from "./Dock.vue";
import { getWinSize } from "@/util/util.js";
import { LinkOutlined, SkinOutlined } from "@vicons/antd";
import { renderIcon } from "@/util/icon";
const LaunchPad = defineAsyncComponent(() => import("./LaunchPad.vue"));
const Window = defineAsyncComponent(() => import("./Window.vue"));
const RightMenu = defineAsyncComponent(() => import("./RightMenu.vue"));
const DesktopIcons = defineAsyncComponent(() => import("./DesktopIcons.vue"));
const DesktopShortcutWin = defineAsyncComponent(() => import("@/components/common/modules/desktop/DesktopShortcutWin.vue"));
import { useEventBus } from "@/util/event.js";
const showTopBarAndDock = ref(true);
const desktopRefreshKey = ref(0);
const ws = $wins;
const personalizeIcon = renderIcon(SkinOutlined, 40);
const shortcutIcon = renderIcon(LinkOutlined, 40);
const refreshDesktop = () => {
  desktopRefreshKey.value++;
};
const isFullscreen = ref(false);
const handleFullscreenChange = () => {
  const ele = document.fullscreenElement || document.webkitFullscreenElement || document.mozFullScreenElement;
  isFullscreen.value = !!ele;
};
const toggleFullscreen = () => {
  if (!isFullscreen.value) {
    const ele = document.documentElement;
    if (ele.requestFullscreen) {
      ele.requestFullscreen();
    } else if (ele.webkitRequestFullscreen) {
      ele.webkitRequestFullscreen();
    } else if (ele.mozRequestFullScreen) {
      ele.mozRequestFullScreen();
    }
  } else if (document.exitFullscreen) {
    document.exitFullscreen();
  } else if (document.webkitExitFullscreen) {
    document.webkitExitFullscreen();
  } else if (document.mozCancelFullScreen) {
    document.mozCancelFullScreen();
  }
};
const openPersonalize = () => {
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width < 740 ? sz.width : 740,
    height: sz.height,
    title: $tc("contextMenu.common.personalize"),
    icon: personalizeIcon,
    component: defineAsyncComponent(
      () => import("@/components/common/modules/settings/PersonalizeTab.vue"),
    ),
    data: {},
  });
};
const openShortcutWin = () => {
  $wins.addWindow({
    width: 520,
    height: 360,
    title: "新建快捷方式",
    icon: shortcutIcon,
    component: DesktopShortcutWin,
    data: {},
  });
};
const desktopContext = computed(() => ({
  type: "macos.desktop",
  actions: {
    refresh: refreshDesktop,
    toggleFullscreen,
    personalize: openPersonalize,
    shortcut: openShortcutWin,
  },
}));
const desktopMenuResolver = () => [
  {
    key: "shortcut",
    label: () => "新建快捷方式",
    group: "main",
    order: 20,
    action: (ctx) => ctx.actions.shortcut(),
  },
  {
    key: "refresh",
    label: () => $t("contextMenu.common.refresh"),
    group: "main",
    order: 10,
    action: (ctx) => ctx.actions.refresh(),
  },
  {
    key: "fullscreen",
    label: () => $t(isFullscreen.value ? "macos.controlCenter.exitFullscreen" : "macos.controlCenter.enterFullscreen"),
    group: "window",
    order: 10,
    action: (ctx) => ctx.actions.toggleFullscreen(),
  },
  {
    key: "personalize",
    label: () => $t("contextMenu.common.personalize"),
    group: "manage",
    order: 10,
    action: (ctx) => ctx.actions.personalize(),
  },
];
$contextMenu.register("macos.desktop", desktopMenuResolver);
onMounted(() => {
  document.addEventListener("fullscreenchange", handleFullscreenChange);
  document.addEventListener("webkitfullscreenchange", handleFullscreenChange);
  document.addEventListener("mozfullscreenchange", handleFullscreenChange);
  handleFullscreenChange();
});
onUnmounted(() => {
  $contextMenu.unregister("macos.desktop", desktopMenuResolver);
  document.removeEventListener("fullscreenchange", handleFullscreenChange);
  document.removeEventListener("webkitfullscreenchange", handleFullscreenChange);
  document.removeEventListener("mozfullscreenchange", handleFullscreenChange);
});
useEventBus("window-status-change", (msg) => {
  showTopBarAndDock.value = $wins.getMaxWindow() == null;
});
const showLaunchPad = ref(false);
useEventBus("launchpad-pad-show", (msg) => {
  showLaunchPad.value = msg;
});
const wallPaper = computed(
  () => {
    const data = {
      DefaultWallpaper: $user.settings.DefaultWallpaper,
      DefaultWallpaperType: $user.settings.DefaultWallpaperType,
    }
    if ($user.settings.DefaultWallpaperType == 'image' && $user.settings.DefaultWallpaper == "default") {
      data.DefaultWallpaper = "/assets/" + $user.settings.Mode + "/img/wallpaper.jpg"
    }
    return data;
  }
)
</script>
<template>
  <div class="w-full h-full overflow-hidden bg-center bg-cover screen-brightness">
    <div class="fixed inset-0 z-0" v-context-menu="desktopContext"></div>
    <img :key="desktopRefreshKey" v-if="
      wallPaper.DefaultWallpaperType == 'image' &&
      wallPaper.DefaultWallpaper
    " class="w-100vw h-100vh object-cover fixed z--1 pointer-events-none" :src="wallPaper.DefaultWallpaper" />
    <video v-if="
      wallPaper.DefaultWallpaperType == 'video' &&
      wallPaper.DefaultWallpaper
    " class="w-100vw h-100vh object-cover fixed z--1 pointer-events-none" :src="wallPaper.DefaultWallpaper" autoplay
      muted loop></video>
    <div class="window-bound z-10 absolute">
      <template v-for="win in ws.windows">
        <Window :data="win" :class="{ hidden: win.close || win.status == 'min' }" />
      </template>
    </div>
    <DesktopIcons :key="desktopRefreshKey" />
    <TopBar v-if="showTopBarAndDock" />
    <Dock v-if="showTopBarAndDock" />
    <LaunchPad v-if="showLaunchPad" />
    <RightMenu v-if="false && showTopBarAndDock" />
  </div>
</template>
<style scoped></style>
