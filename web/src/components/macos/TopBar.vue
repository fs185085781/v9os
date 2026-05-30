<script setup>
import { Apple, Linux, TabletAlt, Windows } from "@vicons/fa";
import { NButton, NIcon, NSlider } from "naive-ui";
import { getWinSize } from "@/util/util.js";
import { WifiOutlined, WifiOffOutlined } from "@vicons/material";
import { SettingOutlined } from "@vicons/antd";
import {
  onMounted,
  defineAsyncComponent,
  computed,
  ref,
  onUnmounted,
  watch,
} from "vue";
import { useEventBus } from "@/util/event.js";
import { elementInMe, delayAction } from "@/util/util.js";
import GlassLayer from "@/components/common/component/util/GlassLayer.vue";
import ChatNotifyPopover from "@/components/common/component/user/chat/ChatNotifyPopover.vue";
import IconView from "@/components/common/component/util/IconView.vue";
import { renderIcon } from "@/util/icon";
const hasAppStore = ref(false);
const settingsIcon = renderIcon(SettingOutlined, 40);
const openMyProfile = async () => {
  leftTools.value.show = false;
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width,
    height: sz.height,
    title: computed(() => {
      return $t("common.settings.title");
    }),
    icon: settingsIcon,
    component: defineAsyncComponent(() => import("./SystemPreferences.vue")),
    data: {},
  });
};
const openAppStore = () => {
  leftTools.value.show = false;
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width,
    height: sz.height,
    title: computed(() => {
      return $t("common.appstore.title");
    }),
    icon: $user.defaultIcon("appstore"),
    component: defineAsyncComponent(() => import("@/components/common/modules/appstore/AppStore.vue")),
    data: {},
  });
};
const rightPanel = ref({
  dateTime: {
    date: "",
    time: "",
  },
  battery: {
    charging: false,
    level: 100,
    width: 100,
  },
  network: true,
});
const rightTools = ref({
  show: false,
});
const leftTools = ref({
  show: false,
});
const leftMenuZIndex = 20;
const calcDateAndOther = (num) => {
  const date = new Date();
  rightPanel.value.dateTime.date = date.format("yyyy-MM-dd");
  rightPanel.value.dateTime.time = date.format("HH:mm:ss");
  if (num % 2 == 0) {
    if (navigator.getBattery) {
      navigator.getBattery().then(function (res) {
        rightPanel.value.battery = {
          charging: res.charging,
          level: Math.floor(res.level * 100),
          width: Math.floor(res.level * 100) * 0.72,
        };
      });
    }else{
      rightPanel.value.battery = {
        charging: false,
        level: 100,
        width: 72
      };
    }
  }
};
calcDateAndOther(0);
let jc = 0;
const timeId = setInterval(() => {
  jc++;
  calcDateAndOther(jc);
  if (jc > 10000) {
    jc = -1;
  }
}, 1000);
onMounted(async () => {
  hasAppStore.value = (await $user.getMyApps()).filter(x => x.code == "__appstore__").length>0;
  document.addEventListener("fullscreenchange", handleFullscreenChange);
  document.addEventListener("webkitfullscreenchange", handleFullscreenChange);
  document.addEventListener("mozfullscreenchange", handleFullscreenChange);
  applyScreenEffects();
});
onUnmounted(() => {
  clearInterval(timeId);
  document.removeEventListener("fullscreenchange", handleFullscreenChange);
  document.removeEventListener("webkitfullscreenchange", handleFullscreenChange);
  document.removeEventListener("mozfullscreenchange", handleFullscreenChange);
});
const rightToolsEl = ref();
const rightToolsBtn = ref();
const leftToolsEl = ref();
const leftToolsBtn = ref();
useEventBus("document-click", (el) => {
  if (
    !elementInMe(rightToolsEl._value, el) &&
    !elementInMe(rightToolsBtn._value, el)
  ) {
    rightTools.value.show = false;
  }
  if (
    !elementInMe(leftToolsEl._value, el) &&
    !elementInMe(leftToolsBtn._value, el)
  ) {
    leftTools.value.show = false;
  }
});
useEventBus("sw-change-status", (data) => {
  rightPanel.value.network = data.mode != "offline";
});
const screenBrightness = ref(Number(localStorage.getItem("v9os.screenBrightness") || 100));
const eyeCareMode = ref(localStorage.getItem("v9os.eyeCareMode") === "true");
const soundVolume = computed({
  get: () => $user.settings.SoundVolume ?? 60,
  set: (value) => $user.setSoundVolume(value),
});
const isDarkMode = computed(() => $user.settings.Theme === "dark");
const isFullscreen = ref(false);
const uiModeItems = [
  {
    labelKey: "macos.controlCenter.uiMode.windows",
    mode: "win10",
    icon: Windows,
  },
  {
    labelKey: "macos.controlCenter.uiMode.linux",
    mode: "deepin",
    icon: Linux,
  },
  {
    labelKey: "macos.controlCenter.uiMode.pad",
    mode: "pad",
    icon: TabletAlt,
  },
];
const applyScreenEffects = () => {
  const root = document.documentElement.style;
  root.setProperty("--screen-brightness", `${screenBrightness.value}%`);
  root.setProperty("--eye-care-sepia", eyeCareMode.value ? "18%" : "0%");
  root.setProperty("--eye-care-saturate", eyeCareMode.value ? "86%" : "100%");
  root.setProperty("--eye-care-warmth", eyeCareMode.value ? "rgba(255, 214, 132, 0.10)" : "transparent");
  localStorage.setItem("v9os.screenBrightness", String(screenBrightness.value));
  localStorage.setItem("v9os.eyeCareMode", String(eyeCareMode.value));
};
watch([screenBrightness, eyeCareMode], applyScreenEffects, { immediate: true });
const toggleTheme = () => {
  $user.setTheme(isDarkMode.value ? "light" : "dark", true);
};
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
const switchUiMode = (mode) => {
  const lastMode = $user.settings.Mode;
  $user.setUiMode(mode, true, lastMode);
};
delayAction(
  () => {
    return !!window.sw;
  },
  () => {
    sw.getMode().then((mode) => {
      if (mode == "offline") {
        rightPanel.value.network = false;
      }
    });
  },
);
const wins = computed(() => {
  return $wins.windows;
});
const openAndTopWin = (w) => {
  if (w.status == "min") {
    $wins.winStatusChange(w, w.lastStatus)
  }
  $wins.activeWindow(w)
}
</script>
<template>
  <div
    class="w-full h-6 px-4 fixed top-0 flex items-center justify-between z-20 text-sm text-white bg-gray-500 bg-opacity-10 backdrop-blur-2xl shadow transition">
    <div class="flex items-center space-x-4">
      <div ref="leftToolsBtn" @click.stop="leftTools.show = !leftTools.show"
        class="cursor-pointer inline-flex h-6 cursor-default flex-row space-x-1 user-rounded-1 hover:user-color-bg p-1">
        <Apple />
      </div>
    </div>
    <GlassLayer
      :visible="leftTools.show"
      :target="leftToolsEl"
      :content-z-index="leftMenuZIndex"
      radius-class="user-rounded-lg"
      fixed
    />
    <div
      ref="leftToolsEl"
      v-if="leftTools.show"
      class="menu-box top-6 left-4 w-56 user-rounded-lg v9os-glass-surface"
      :style="{ zIndex: leftMenuZIndex }"
    >
      <ul class="py-1 border-b user-color-border">
        <li class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="openMyProfile()">
          {{ $t("common.settings.title") }}
        </li>
        <li v-if="hasAppStore" class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="openAppStore()">
          {{ $t("common.appstore.title") }}
        </li>
      </ul>
      <ul class="py-1 border-b user-color-border">
        <li class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="$user.logout()">
          {{ $t("common.all.logout") }}
        </li>
      </ul>
      <ul class="py-1">
        <li class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="
            $user.system.Wakeup = true;
          $user.system.Shutdown = true;
          ">
          {{ $t("common.all.sleep") }}
        </li>
        <li class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="
            $user.system.Open = true;
          $user.system.Shutdown = true;
          ">
          {{ $t("common.all.restart") }}
        </li>
        <li class="px-5 leading-6 cursor-default user-color-ftext hover:user-color-bg" @click="$user.system.Shutdown = true">
          {{ $t("common.all.shutdown") }}
        </li>
      </ul>
    </div>
    <div class="flex items-center justify-end space-x-1">
      <div class="flex items-center space-x-0.5">
        <template v-for="item in wins" :key="item.id">
          <div v-if="!item.close && !item.parentId"
            class="relative group cursor-pointer px-1.5 py-0.5 user-rounded-1 transition-all duration-200 hover:user-color-bg"
            :title="item.title" @click="openAndTopWin(item)">
            <IconView :icon="item.icon" :size="18" icon-class="w-4.5 h-4.5 object-contain transition-transform duration-200 group-hover:scale-110" />
          </div>
        </template>
      </div>
      <div
        class="hidden sm:inline-flex h-6 cursor-default flex-row space-x-1 user-rounded-1 hover:user-color-bg p-1">
        <div class="flex-center">
          <span class="text-xs mt-0.5 mr-2">{{ rightPanel.battery.level }}%</span>
          <div class="relative">
            <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="24" width="24"
              xmlns="http://www.w3.org/2000/svg">
              <path
                d="M0 6a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V6zm2-1a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V6a1 1 0 0 0-1-1H2zm14 3a1.5 1.5 0 0 1-1.5 1.5v-3A1.5 1.5 0 0 1 16 8z">
              </path>
            </svg>
            <div :class="{
              'bg-green-400': rightPanel.battery.charging,
              'user-color-fbg': !rightPanel.battery.charging,
            }" class="battery-level" :style="{ width: rightPanel.battery.width + '%' }"></div>
            <svg v-if="rightPanel.battery.charging" stroke="currentColor" fill="currentColor" stroke-width="0"
              viewBox="0 0 16 16" class="absolute top-1/2 -mt-1.5 left-0 ml-1" height="12" width="12"
              xmlns="http://www.w3.org/2000/svg">
              <path
                d="M11.251.068a.5.5 0 0 1 .227.58L9.677 6.5H13a.5.5 0 0 1 .364.843l-8 8.5a.5.5 0 0 1-.842-.49L6.323 9.5H3a.5.5 0 0 1-.364-.843l8-8.5a.5.5 0 0 1 .615-.09z">
              </path>
            </svg>
          </div>
        </div>
      </div>
      <div
        class="hidden sm:inline-flex h-6 cursor-default flex-row space-x-1 user-rounded-1 hover:user-color-bg p-1">
        <n-icon :component="rightPanel.network ? WifiOutlined : WifiOffOutlined" size="20" class="m--0.5" />
      </div>
      <ChatNotifyPopover placement="bottom-end" trigger-class="relative inline-flex h-6 cursor-pointer flex-row items-center user-rounded-1 hover:user-color-bg p-1" />
      <div ref="rightToolsBtn"
        class="inline-flex h-6 cursor-default flex-row space-x-1 user-rounded-1 hover:user-color-bg p-1 cursor-pointer"
        @click.stop="rightTools.show = !rightTools.show">
        <svg viewBox="0 0 29 29" width="16" height="16" xmlns="http://www.w3.org/2000/svg" fill="currentColor">
          <path
            d="M7.5,13h14a5.5,5.5,0,0,0,0-11H7.5a5.5,5.5,0,0,0,0,11Zm0-9h14a3.5,3.5,0,0,1,0,7H7.5a3.5,3.5,0,0,1,0-7Zm0,6A2.5,2.5,0,1,0,5,7.5,2.5,2.5,0,0,0,7.5,10Zm14,6H7.5a5.5,5.5,0,0,0,0,11h14a5.5,5.5,0,0,0,0-11Zm1.43439,8a2.5,2.5,0,1,1,2.5-2.5A2.5,2.5,0,0,1,22.93439,24Z">
          </path>
        </svg>
      </div>
      <div ref="rightToolsEl" v-if="rightTools.show"
        class="control-center user-rounded-2xl grid fixed shadow-base w-84 max-w-full top-8 right-0 sm:right-2 p-2.5 user-color-ftext user-color-fbg border user-color-border grid-cols-4 gap-2">
        <div class="cc-grid row-span-2 col-span-2 p-2 flex flex-col justify-around">
          <div v-for="item in uiModeItems" :key="item.mode" class="flex-center-v space-x-2 cursor-pointer user-rounded-lg px-1 py-1"
            :class="{ 'user-color-bg': $user.settings.Mode === item.mode }" @click="switchUiMode(item.mode)">
            <span class="cc-system-icon">
              <n-icon :component="item.icon" />
            </span>
            <div class="flex flex-col pt-0.5">
              <span class="font-medium leading-4">{{ $t(item.labelKey) }}</span>
            </div>
          </div>
        </div>
        <div class="cc-grid col-span-2 p-2 flex-center-v space-x-2 cursor-pointer" @click="toggleTheme">
          <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 512 512" class="cc-mode"
            height="32" width="32" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M256 118a22 22 0 01-22-22V48a22 22 0 0144 0v48a22 22 0 01-22 22zm0 368a22 22 0 01-22-22v-48a22 22 0 0144 0v48a22 22 0 01-22 22zm113.14-321.14a22 22 0 01-15.56-37.55l33.94-33.94a22 22 0 0131.11 31.11l-33.94 33.94a21.93 21.93 0 01-15.55 6.44zM108.92 425.08a22 22 0 01-15.55-37.56l33.94-33.94a22 22 0 1131.11 31.11l-33.94 33.94a21.94 21.94 0 01-15.56 6.45zM464 278h-48a22 22 0 010-44h48a22 22 0 010 44zm-368 0H48a22 22 0 010-44h48a22 22 0 010 44zm307.08 147.08a21.94 21.94 0 01-15.56-6.45l-33.94-33.94a22 22 0 0131.11-31.11l33.94 33.94a22 22 0 01-15.55 37.56zM142.86 164.86a21.89 21.89 0 01-15.55-6.44l-33.94-33.94a22 22 0 0131.11-31.11l33.94 33.94a22 22 0 01-15.56 37.55zM256 358a102 102 0 11102-102 102.12 102.12 0 01-102 102z">
            </path>
          </svg>
          <div class="flex flex-col">
            <span class="font-medium ml-1">{{ isDarkMode ? $t("macos.controlCenter.darkMode") : $t("macos.controlCenter.lightMode") }}</span>
            <span class="cc-text ml-1">{{ $t("macos.controlCenter.clickToSwitch") }}</span>
          </div>
        </div>
        <div class="cc-grid flex-center flex-col text-center cursor-pointer" :class="{ 'active-tile': eyeCareMode }"
          @click="eyeCareMode = !eyeCareMode">
          <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="20" width="20"
            xmlns="http://www.w3.org/2000/svg">
            <path
              d="M8 3a.5.5 0 0 1 .5.5v2a.5.5 0 0 1-1 0v-2A.5.5 0 0 1 8 3zm8 8a.5.5 0 0 1-.5.5h-2a.5.5 0 0 1 0-1h2a.5.5 0 0 1 .5.5zm-13.5.5a.5.5 0 0 0 0-1h-2a.5.5 0 0 0 0 1h2zm11.157-6.157a.5.5 0 0 1 0 .707l-1.414 1.414a.5.5 0 1 1-.707-.707l1.414-1.414a.5.5 0 0 1 .707 0zm-9.9 2.121a.5.5 0 0 0 .707-.707L3.05 5.343a.5.5 0 1 0-.707.707l1.414 1.414zM8 7a4 4 0 0 0-4 4 .5.5 0 0 0 .5.5h7a.5.5 0 0 0 .5-.5 4 4 0 0 0-4-4zm0 1a3 3 0 0 1 2.959 2.5H5.04A3 3 0 0 1 8 8z">
            </path>
          </svg><span class="text-xs leading-cc">{{ $t("macos.controlCenter.eyeCare") }}</span>
          <span class="cc-text">{{ eyeCareMode ? $t("macos.controlCenter.on") : $t("macos.controlCenter.off") }}</span>
        </div>
        <div class="cc-grid flex-center flex-col text-center cursor-pointer" @click="toggleFullscreen">
          <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="16" width="16"
            xmlns="http://www.w3.org/2000/svg">
            <path
              d="M1.5 1a.5.5 0 0 0-.5.5v4a.5.5 0 0 1-1 0v-4A1.5 1.5 0 0 1 1.5 0h4a.5.5 0 0 1 0 1h-4zM10 .5a.5.5 0 0 1 .5-.5h4A1.5 1.5 0 0 1 16 1.5v4a.5.5 0 0 1-1 0v-4a.5.5 0 0 0-.5-.5h-4a.5.5 0 0 1-.5-.5zM.5 10a.5.5 0 0 1 .5.5v4a.5.5 0 0 0 .5.5h4a.5.5 0 0 1 0 1h-4A1.5 1.5 0 0 1 0 14.5v-4a.5.5 0 0 1 .5-.5zm15 0a.5.5 0 0 1 .5.5v4a1.5 1.5 0 0 1-1.5 1.5h-4a.5.5 0 0 1 0-1h4a.5.5 0 0 0 .5-.5v-4a.5.5 0 0 1 .5-.5z">
            </path>
          </svg><span class="text-xs leading-cc mt-1.5">{{ isFullscreen ? $t("macos.controlCenter.exitFullscreen") : $t("macos.controlCenter.enterFullscreen") }}</span>
        </div>
        <div class="cc-grid col-span-4 px-2.5 py-2 flex flex-col justify-around">
          <div class="flex justify-between">
            <span class="font-medium ml-0.5">{{ $t("macos.controlCenter.display") }}</span>
            <span class="cc-text">{{ screenBrightness }}%</span>
          </div>
          <div class="slider flex w-full">
            <n-slider v-model:value="screenBrightness" :min="45" :max="125" :step="1" :tooltip="true" />
          </div>
        </div>
        <div class="cc-grid col-span-4 px-2.5 py-2 flex flex-col justify-around">
          <div class="flex justify-between">
            <span class="font-medium ml-0.5">{{ $t("macos.controlCenter.sound") }}</span>
            <span class="cc-text">{{ soundVolume }}%</span>
          </div>
          <div class="slider flex w-full">
            <n-slider v-model:value="soundVolume" :min="0" :max="100" :step="1" :tooltip="true" />
          </div>
        </div>
      </div>
      <span>{{ rightPanel.dateTime.date }}</span><span>{{ rightPanel.dateTime.time }}</span>
    </div>
  </div>
</template>
