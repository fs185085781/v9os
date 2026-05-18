<script setup>
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref } from "vue";
import { NIcon } from "naive-ui";
import {
  ChevronBack,
  ChevronForward,
  Power,
  VolumeHigh,
} from "@vicons/ionicons5";
import { KeyboardOutlined, WifiOffOutlined, WifiOutlined } from "@vicons/material";
import { delayAction, getWinSize } from "@/util/util.js";
import { useStore } from "@/stores/user.js";
import { useEventBus } from "@/util/event.js";
import ChatNotifyPopover from "@/components/common/component/user/chat/ChatNotifyPopover.vue";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/IconView.vue";

const store = useStore();
const LaunchPad = defineAsyncComponent(() => import("./LaunchPad.vue"));
const moduleInfoMap = ref({});
const drawerOpen = ref(true);
const dateTime = ref({
  time: "",
  date: "",
});
const rightPanel = ref({
  battery: {
    charging: false,
    level: 100,
    width: 72,
  },
  network: true,
});

const formatDate = () => {
  const date = new Date();
  dateTime.value = {
    time: `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`,
    date: `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`,
  };
};

const calcBattery = () => {
  if (navigator.getBattery) {
    navigator.getBattery().then((res) => {
      rightPanel.value.battery = {
        charging: res.charging,
        level: Math.floor(res.level * 100),
        width: Math.floor(res.level * 100) * 0.72,
      };
    });
  } else {
    rightPanel.value.battery = {
      charging: false,
      level: 100,
      width: 72,
    };
  }
};

formatDate();
calcBattery();
let tickCount = 0;
const timer = setInterval(() => {
  tickCount++;
  formatDate();
  if (tickCount % 2 == 0) {
    calcBattery();
  }
}, 1000);

async function loadDynamicApps() {
  const map = {};
  const userApps = await $user.getMyApps();
  if (userApps) {
    userApps.forEach((mod) => {
      map[mod.code] = mod;
    });
  }
  moduleInfoMap.value = map;
}

onMounted(() => {
  loadDynamicApps();
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
});
onUnmounted(() => {
  clearInterval(timer);
});

useEventBus("sw-change-status", (data) => {
  rightPanel.value.network = data.mode != "offline";
});

const dynamicApps = computed(() => {
  const dockCodes = $user.getDockApps ? $user.getDockApps() : [];
  return dockCodes
    .map((code) => moduleInfoMap.value[code])
    .filter(Boolean);
});

const apps = computed(() => dynamicApps.value);
const wins = computed(() => $wins.windows.filter((win) => !win.close && !win.parentId));
const batteryTitle = computed(() => {
  const level = rightPanel.value.battery.level;
  if (rightPanel.value.battery.charging) {
    return $t("common.all.batteryCharging", { level });
  }
  return $t("common.all.batteryAvailable", { level });
});

const openApp = (app) => {
  if (app.type === "noop") return;
  const sz = getWinSize();
  const options = {
    id: app.windowId || "",
    width: sz.width,
    height: sz.height - 78,
    title: app.name,
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

const openAndTopWin = (w) => {
  if (w.status == "min") {
    $wins.winStatusChange(w, w.lastStatus)
  }
  $wins.activeWindow(w);
};

const openLaunchPad = () => {
  const sz = getWinSize();
  $wins.addWindow({
    id: "__deepin_launchpad__",
    width: Math.min(sz.width - 80, 920),
    height: Math.min(sz.height - 120, 640),
    title: "启动器",
    icon: "/assets/deepin/img/icons/menu.png",
    component: LaunchPad,
    data: {},
  });
};

const logoutAccount = async () => {
  if (await $msg.util.confirm($t("common.all.logoutConfirm"))) {
    $user.logout();
  }
};

const removeFromDock = (app) => {
  if (store.removeDockApp && app.code) {
    store.removeDockApp(app.code);
  }
};

const appContext = (app) => ({
  type: "deepin.dock.app",
  payload: app,
  actions: {
    addToDesktop: () => openDesktopShortcutWin(app, { title: "添加桌面快捷方式" }),
    removeFromDock: () => removeFromDock(app),
  },
});

const appMenuResolver = () => [
  {
    key: "addToDesktop",
    label: () => "添加到桌面",
    group: "main",
    order: 10,
    visible: (ctx) => ctx.payload.type !== "noop",
    action: (ctx) => ctx.actions.addToDesktop(),
  },
  {
    key: "removeFromDock",
    label: () => $t("contextMenu.common.removeFromDock"),
    group: "main",
    order: 20,
    visible: (ctx) => ctx.payload.type !== "noop",
    action: (ctx) => ctx.actions.removeFromDock(),
  },
];

$contextMenu.register("deepin.dock.app", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("deepin.dock.app", appMenuResolver);
});
</script>

<template>
  <div class="fixed left-0 right-0 bottom-0 z-50 min-w-150 p-3">
    <div class="relative h-12 min-w-0 flex items-center justify-center backdrop-blur-30 bg-[color-mix(in_srgb,var(--user-bg-color)_52%,transparent)] shadow-[0_4px_10px_color-mix(in_srgb,var(--user-text-color)_30%,transparent)] user-rounded-3">
      <div class="absolute left-2 top-0 h-full flex items-center">
        <button class="w-10 h-10 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" title="Starter" @click="openLaunchPad">
          <img class="w-8 h-8 object-contain" src="/assets/deepin/img/icons/menu.png" alt="" />
        </button>
      </div>

      <div class="h-full max-w-[min(52vw,720px)] max-[1124px]:max-w-[34vw] overflow-hidden flex items-center justify-center gap-3">
        <button
          v-for="app in apps.filter((item) => item.type !== 'noop')"
          :key="app.code"
          class="w-10 h-10 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]"
          :title="typeof app.name === 'string' ? app.name : ''"
          @click="openApp(app)"
          v-context-menu="appContext(app)"
        >
          <IconView :icon="app.icon" :size="32" icon-class="w-8 h-8 object-contain" />
        </button>
        <button
          v-for="win in wins"
          :key="`win-${win.id}`"
          class="w-10 h-10 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)] after:content-[''] after:absolute after:bottom-0.25 after:left-1/2 after:-translate-x-1/2 after:w-1 after:h-1 after:bg-[var(--user-text-color)]"
          :class="{ 'after:block': win.active && win.status !== 'min', 'after:hidden': !(win.active && win.status !== 'min') }"
          :title="win.title"
          @click="openAndTopWin(win)"
        >
          <IconView :icon="win.icon" fallback="/assets/deepin/img/icons/app.png" :size="32" icon-class="w-8 h-8 object-contain" />
        </button>
      </div>

      <div class="absolute right-2 top-0 h-full flex items-center">
        <ChatNotifyPopover
          placement="top-end"
          :icon-size="22"
          trigger-class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)] border-0 bg-transparent"
        />
        <div class="h-11 min-w-20 flex flex-col items-center justify-center user-color-ftext">
          <strong class="text-5 leading-5.5">{{ dateTime.time }}</strong>
          <span class="text-3 font-700 leading-3.5">{{ dateTime.date }}</span>
        </div>
        <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" :title="$t('common.all.logout')" @click="logoutAccount">
          <n-icon :component="Power" size="22" />
        </button>
        <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" title="On-Screen Keyboard">
          <n-icon :component="KeyboardOutlined" size="22" />
        </button>
        <div class="w-0.5 h-7.5 mx-2 bg-[color-mix(in_srgb,var(--user-text-color)_26%,transparent)]"></div>
        <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" title="Folder" @click="drawerOpen = !drawerOpen">
          <n-icon :component="drawerOpen ? ChevronForward : ChevronBack" size="20" />
        </button>
        <div class="flex items-center gap-1 overflow-hidden transition-[max-width,opacity] duration-200 max-[1124px]:max-w-0 max-[1124px]:opacity-0"
          :class="{ 'max-w-32.5 opacity-100': drawerOpen, 'max-w-0 opacity-0': !drawerOpen }">
          <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" :title="batteryTitle">
            <div class="relative w-6 h-6 flex items-center justify-center">
              <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="24" width="24"
                xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M0 6a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V6zm2-1a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V6a1 1 0 0 0-1-1H2zm14 3a1.5 1.5 0 0 1-1.5 1.5v-3A1.5 1.5 0 0 1 16 8z">
                </path>
              </svg>
              <div
                class="absolute left-0.5 top-2 h-2 max-w-[72%] user-rounded-sm"
                :class="{
                  'bg-green-400': rightPanel.battery.charging,
                  'bg-[var(--user-text-color)]': !rightPanel.battery.charging,
                }"
                :style="{ width: rightPanel.battery.width + '%' }"
              ></div>
              <svg v-if="rightPanel.battery.charging" stroke="currentColor" fill="currentColor" stroke-width="0"
                viewBox="0 0 16 16" class="absolute left-1 top-1.5" height="12" width="12"
                xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M11.251.068a.5.5 0 0 1 .227.58L9.677 6.5H13a.5.5 0 0 1 .364.843l-8 8.5a.5.5 0 0 1-.842-.49L6.323 9.5H3a.5.5 0 0 1-.364-.843l8-8.5a.5.5 0 0 1 .615-.09z">
                </path>
              </svg>
            </div>
          </button>
          <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" title="Network">
            <n-icon :component="rightPanel.network ? WifiOutlined : WifiOffOutlined" size="22" />
          </button>
          <button class="w-8.5 h-8.5 flex flex-none items-center justify-center relative user-color-ftext user-rounded-2 transition-[background,transform] duration-150 hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]" title="Volume">
            <n-icon :component="VolumeHigh" size="22" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
