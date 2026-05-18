<script setup>
import {
  onMounted,
  onUnmounted,
  computed,
  ref
} from "vue";
import { getWinSize} from "../../util/util.js";
import { useStore } from "@/stores/user.js";
import emitter, { useEventBus } from "@/util/event.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/IconView.vue";
const store = useStore();
let animationFrameId = null;
let lastMouseX = null;
let isMouseOver = false;
const dockIconSizes = ref({});
const dockIconSize = (app) => dockIconSizes.value[app.code] || 50;

const handleMouseMove = (event) => {
  lastMouseX = event.clientX;
  isMouseOver = true;
  if (!animationFrameId) {
    animationFrameId = requestAnimationFrame(updateIconSizes);
  }
};
const handleMouseLeave = () => {
  isMouseOver = false;
  lastMouseX = null;

  const icons = document.querySelectorAll(".dock li .dock-icon-visual");
  icons.forEach((icon) => {
    icon.style.width = "50px";
  });
  dockIconSizes.value = {};

  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
};

const updateIconSizes = () => {
  if (!isMouseOver || lastMouseX === null) {
    animationFrameId = null;
    return;
  }

  const icons = document.querySelectorAll(".dock li .dock-icon-visual");
  const baseSize = 50;
  const maxSize = 90;
  const distanceLimit = 300;
  const nextSizes = {};

  icons.forEach((icon, index) => {
    const rect = icon.getBoundingClientRect();
    const imgCenterX = rect.left + rect.width / 2;
    const distance = Math.abs(lastMouseX - imgCenterX);
    const app = apps.value[index];

    if (distance > distanceLimit) {
      icon.style.width = `${baseSize}px`;
      if (app) {
        nextSizes[app.code] = baseSize;
      }
    } else {
      const normalizedDistance = distance / distanceLimit;
      const scale = 1 + (1 - Math.pow(normalizedDistance, 0.7)) * 0.8;
      const newWidth = Math.min(baseSize * scale, maxSize);
      icon.style.width = `${newWidth}px`;
      if (app) {
        nextSizes[app.code] = newWidth;
      }
    }
  });
  dockIconSizes.value = nextSizes;

  animationFrameId = requestAnimationFrame(updateIconSizes);
};

onMounted(() => {
  const dockElement = document.querySelector(".dock");
  if (dockElement) {
    dockElement.addEventListener("mousemove", handleMouseMove);
    dockElement.addEventListener("mouseleave", handleMouseLeave);
  }
  loadDynamicApps();
});

onUnmounted(() => {
  const dockElement = document.querySelector(".dock");
  if (dockElement) {
    dockElement.removeEventListener("mousemove", handleMouseMove);
    dockElement.removeEventListener("mouseleave", handleMouseLeave);
  }

  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
  }
});
const fixedApps = [
  {
    code: "__launchpad__",
    name: computed(() => $t("macos.app.launchpad")),
    url: "launchpad",
    type: "pad",
    icon: "/assets/macos/img/icons/launchpad.png",
  }
];
const moduleInfoMap = ref({});
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

useEventBus("dock-apps-change", () => {
  loadDynamicApps();
});
const dynamicApps = computed(() => {
  const dockCodes = $user.getDockApps ? $user.getDockApps() : [];
  return dockCodes
    .map((code) => {
      const info = moduleInfoMap.value[code];
      if (info) {
        return info;
      }
      return null;
    })
    .filter(Boolean);
});
const apps = computed(() => [...fixedApps, ...dynamicApps.value]);
const openApp = (app) => {
  if (app.type == "pad") {
    emitter.emit(app.url + "-pad-show", true);
    return;
  }
  const sz = getWinSize();
  const options = {
    id: app.windowId || "",
    width: sz.width,
    height: sz.height,
    title: app.name,
    icon:app.icon,
    data: {},
  };
  if (app.type === "system") {
    options.component = app.url;
  } else if (app.type === "plugin" || app.type === "url") {
    options.iframeUrl = app.url;
  }
  $wins.addWindow(options);
};

const removeFromDock = (app) => {
  if (store.removeDockApp && app.code) {
    store.removeDockApp(app.code);
  }
};
const appContext = (app) => ({
  type: "macos.dock.app",
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
    visible: (ctx) => ctx.payload.type !== "pad",
    action: (ctx) => ctx.actions.addToDesktop(),
  },
  {
    key: "removeFromDock",
    label: () => $t("contextMenu.common.removeFromDock"),
    group: "main",
    order: 20,
    visible: (ctx) => ctx.payload.type !== "pad",
    action: (ctx) => ctx.actions.removeFromDock(),
  },
];
$contextMenu.register("macos.dock.app", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("macos.dock.app", appMenuResolver);
});
</script>

<template>
  <div class="dock w-full sm:w-max fixed left-0 right-0 mx-auto bottom-0 z-50 overflow-x-scroll sm:overflow-x-visible">
    <ul class="macos-dock-panel user-rounded-2 mx-auto w-max px-2 space-x-2 flex"
      style="height: 65px">
      <li v-for="app in apps" :key="app.code"
        class="flex-center-v flex-col justify-end mb-1 transition duration-150 ease-in origin-bottom"
        @click="openApp(app)" v-context-menu="appContext(app)">
        <p class="tooltip absolute px-3 py-1 user-rounded-md text-sm user-color-fbg user-color-ftext">
          {{ app.name }}
        </p>
        <span
          class="dock-icon-visual flex items-center justify-center"
          :title="typeof app.name === 'string' ? app.name : ''"
          draggable="false"
          style="will-change: width; width: 50px"
        >
          <IconView :icon="app.icon" :alt="typeof app.name === 'string' ? app.name : ''" :size="dockIconSize(app)" icon-class="w-full object-contain" />
        </span>
        <div class="h-1 w-1 m-0 user-rounded-full user-color-ftext invisible"></div>
      </li>
    </ul>
  </div>
</template>
