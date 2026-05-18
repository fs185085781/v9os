<script setup>
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref } from "vue";
import { Windows } from "@vicons/fa";
import { NIcon } from "naive-ui";
import { WifiOutlined, WifiOffOutlined } from "@vicons/material";
import { useEventBus } from "@/util/event.js";
import { elementInMe, delayAction } from "@/util/util.js";
import emitter from "@/util/event.js";
import ChatNotifyPopover from "@/components/common/component/user/chat/ChatNotifyPopover.vue";
import IconView from "@/components/common/IconView.vue";

const StartMenu = defineAsyncComponent(() => import("./StartMenu.vue"));
const startBtn = ref();
const startMenuEl = ref();
const startMenuShow = ref(false);
const rightPanel = ref({
  dateTime: {
    date: "",
    time: "",
  },
  battery: {
    charging: false,
    level: 100,
    width: 72,
  },
  network: true,
});

const calcDateAndBattery = (num) => {
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
    } else {
      rightPanel.value.battery = {
        charging: false,
        level: 100,
        width: 72,
      };
    }
  }
};
calcDateAndBattery(0);
let jc = 0;
const timeId = setInterval(() => {
  jc++;
  calcDateAndBattery(jc);
  if (jc > 10000) {
    jc = -1;
  }
}, 1000);

onMounted(() => {
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
  clearInterval(timeId);
});

const batteryTitle = computed(() => {
  const level = rightPanel.value.battery.level;
  const charging = rightPanel.value.battery.charging;
  if (charging) {
    return `电池状态：${level}%可用(电源已接通)`;
  }
  return `电池状态：${level}%可用`;
});

useEventBus("sw-change-status", (data) => {
  rightPanel.value.network = data.mode != "offline";
});

useEventBus("document-click", (el) => {
  if (
    !elementInMe(startMenuEl._value, el) &&
    !elementInMe(startBtn._value, el)
  ) {
    emitter.emit("win10-start-show", false);
  }
});

useEventBus("win10-start-show", (msg) => {
  startMenuShow.value = msg;
});

const wins = computed(() => {
  return $wins.windows;
});

const openAndTopWin = (w) => {
  if (w.status == "min") {
    $wins.winStatusChange(w, w.lastStatus)
  }
  $wins.activeWindow(w);
};

const toggleStart = () => {
  emitter.emit("win10-start-show", !startMenuShow.value);
};
</script>

<template>
  <div class="win10-taskbar fixed left-0 right-0 bottom-0 z-50 h-10">
    <div class="flex h-full items-center justify-between">
      <div class="flex h-full items-center">
        <button ref="startBtn" class="win10-taskbar-btn w-12" @click.stop="toggleStart">
          <n-icon :component="Windows" size="18" />
        </button>
        <div class="flex h-full items-center">
          <template v-for="item in wins" :key="item.id">
            <button
              v-if="!item.close && !item.parentId"
              class="win10-taskbar-app"
              :class="{ active: item.active && item.status !== 'min' }"
              :title="item.title"
              @click="openAndTopWin(item)"
            >
              <IconView :icon="item.icon" :size="24" icon-class="w-6 h-6 object-contain" />
            </button>
          </template>
        </div>
      </div>
      <div class="flex h-full items-center space-x-2 px-2">
        <ChatNotifyPopover
          placement="top-end"
          :icon-size="18"
          trigger-class="win10-taskbar-tray-item flex items-center justify-center user-rounded-1 px-2 py-2.7 cursor-pointer border-0 bg-transparent"
        />
        <div class="win10-taskbar-tray-item flex items-center justify-center user-rounded-1 px-2 py-2.7 cursor-default" style="min-width: 32px; min-height: 32px;">
          <n-icon :component="rightPanel.network ? WifiOutlined : WifiOffOutlined" size="18" />
        </div>
        <div class="win10-taskbar-tray-item flex items-center justify-center user-rounded-1 px-1.5 py-2 cursor-default" style="min-height: 32px;" :title="batteryTitle">
          <div class="flex-center">
            <div class="relative">
              <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="24" width="24"
                xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M0 6a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V6zm2-1a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V6a1 1 0 0 0-1-1H2zm14 3a1.5 1.5 0 0 1-1.5 1.5v-3A1.5 1.5 0 0 1 16 8z">
                </path>
              </svg>
              <div :class="{
                'bg-green-400': rightPanel.battery.charging,
                'win10-battery-level-normal': !rightPanel.battery.charging,
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
        <div class="win10-clock win10-taskbar-tray-item h-full flex flex-col items-end justify-center text-xs leading-4 user-rounded-1">
          <span class="text-center w-full">{{ rightPanel.dateTime.time }}</span>
          <span class="text-center w-full">{{ rightPanel.dateTime.date }}</span>
        </div>
      </div>
    </div>
    <div ref="startMenuEl">
      <StartMenu v-if="startMenuShow" />
    </div>
  </div>
</template>

<style scoped>
.battery-level {
  position: absolute;
  top: 13px;
  left: 0px;
  height: 6px;
  border-radius: 1px;
}
.flex-center {
  display: flex;
  align-items: center;
}
</style>
