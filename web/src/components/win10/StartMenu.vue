<script setup>
import { defineAsyncComponent, onMounted, onUnmounted, ref, computed } from "vue";
import { getWinSize } from "@/util/util.js";
import { useStore } from "@/stores/user.js";
import emitter from "@/util/event.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/IconView.vue";
import { SettingOutlined, SkinOutlined, UserOutlined } from "@vicons/antd";
import { renderIcon } from "@/util/icon";

const store = useStore();
const myApps = ref([]);
const userMenuShow = ref(false);
const profileIcon = renderIcon(UserOutlined, 40);
const personalizeIcon = renderIcon(SkinOutlined, 40);
const settingsIcon = renderIcon(SettingOutlined, 40);

onMounted(async () => {
  myApps.value = await $user.getMyApps();
});

const visibleApps = computed(() => {
  return myApps.value;
});

const pinnedApps = computed(() => {
  const dockCodes = $user.getDockApps ? $user.getDockApps() : [];
  return dockCodes
    .map((code) => visibleApps.value.find((app) => app.code === code))
    .filter(Boolean);
});

const appName = (app) => {
  return app.name?.value || app.name || app.code;
};

const closeStartMenu = () => {
  emitter.emit("win10-start-show", false);
};

const openApp = (app) => {
  closeStartMenu();
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

const addToDock = (app) => {
  if (store.addDockApp) {
    store.addDockApp(app.code);
  }
};
const addToDesktop = (app) => {
  openDesktopShortcutWin(app, { title: "添加桌面快捷方式" });
};

const removeFromDock = (app) => {
  if (store.removeDockApp && app.code) {
    store.removeDockApp(app.code);
  }
};
const appContext = (app) => ({
  type: "win10.startMenu.app",
  payload: app,
  actions: {
    pin: () => addToDock(app),
    desktop: () => addToDesktop(app),
    unpin: () => removeFromDock(app),
  },
});
const pinnedContext = (app) => ({
  type: "win10.startMenu.pinned",
  payload: app,
  actions: {
    desktop: () => addToDesktop(app),
    unpin: () => removeFromDock(app),
  },
});
const appMenuResolver = () => [
  {
    key: "desktop",
    label: () => "添加到桌面",
    group: "main",
    order: 20,
    action: (ctx) => ctx.actions.desktop(),
  },
  {
    key: "pin",
    label: () => $t("contextMenu.common.addToMyApps"),
    group: "main",
    order: 10,
    visible: (ctx) => !$user.getDockApps().includes(ctx.payload.code),
    action: (ctx) => {
      ctx.actions.pin();
    },
  },
];
const pinnedMenuResolver = () => [
  {
    key: "desktop",
    label: () => "添加到桌面",
    group: "main",
    order: 10,
    action: (ctx) => ctx.actions.desktop(),
  },
  {
    key: "unpin",
    label: () => $t("contextMenu.common.removeFromMyApps"),
    group: "main",
    order: 20,
    action: (ctx) => {
      ctx.actions.unpin();
    },
  },
];
$contextMenu.register("win10.startMenu.app", appMenuResolver);
$contextMenu.register("win10.startMenu.pinned", pinnedMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("win10.startMenu.app", appMenuResolver);
  $contextMenu.unregister("win10.startMenu.pinned", pinnedMenuResolver);
});

const openProfileTab = () => {
  userMenuShow.value = false;
  closeStartMenu();
  const sz = getWinSize();
  $wins.addWindow({
     width: sz.width<740?sz.width:740,
    height: sz.height,
    title: $tc("backend.layout.profile"),
    icon: profileIcon,
    component: defineAsyncComponent(
      () => import("@/components/common/modules/settings/ProfileTab.vue"),
    ),
    data: {},
  });
};

const openPersonalizeTab = () => {
  userMenuShow.value = false;
  closeStartMenu();
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width<740?sz.width:740,
    height: sz.height,
    title: $tc("backend.layout.settings"),
    icon: personalizeIcon,
    component: defineAsyncComponent(
      () => import("@/components/common/modules/settings/PersonalizeTab.vue"),
    ),
    data: {},
  });
};

const openSettings = () => {
  closeStartMenu();
  $wins.addWindow({
    width: 740,
    height: 600,
    title: $tc("common.settings.title"),
    icon: settingsIcon,
    component: defineAsyncComponent(
      () => import("./SystemPreferences.vue"),
    ),
    data: {},
  });
};

const logout = () => {
  userMenuShow.value = false;
  closeStartMenu();
  $user.logout();
};

const shutdown = () => {
  closeStartMenu();
  $user.system.Shutdown = true;
};
</script>

<template>
  <div class="win10-start-menu user-rounded-t-3 fixed left-0 bottom-10 z-60 h-[640px] max-h-[calc(100vh-2.5rem)] w-[640px] max-w-full bg-black/82 text-gray-100 shadow-2xl backdrop-blur">
    <div class="flex h-full">
      <div class="w-12 flex flex-col justify-between py-2">
        <button class="win10-start-side-btn">
          <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" height="16" width="16">
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <div class="flex flex-col">
          <div class="relative">
            <div v-if="userMenuShow" class="win10-user-menu">
              <button class="win10-user-menu-item" @click="openProfileTab">
                {{ $t("win10.startMenu.user.profile") }}
              </button>
              <button class="win10-user-menu-item" @click="openPersonalizeTab">
                {{ $t("win10.startMenu.user.accountSettings") }}
              </button>
              <button class="win10-user-menu-item" @click="logout">
                {{ $t("win10.startMenu.user.logout") }}
              </button>
            </div>
          <button class="win10-start-side-btn" :title="$user.user.Name || $user.user.Username" @click.stop="userMenuShow = !userMenuShow">
            <img v-if="$user.user.Avatar" :src="$user.user.Avatar" class="w-7 h-7 rounded-full object-cover" />
            <span v-else class="w-7 h-7 rounded-full user-color-bg flex-center font-600">
              {{ $user.user.Username?.charAt(0)?.toUpperCase() }}
            </span>
          </button>
          </div>
          <button class="win10-start-side-btn" :title="$t('win10.startMenu.actions.settings')" @click="openSettings">
            <svg stroke="currentColor" fill="none" stroke-width="1.8" viewBox="0 0 24 24" height="16" width="16">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.5 1.2V21a2 2 0 0 1-4 0v-.09A1.7 1.7 0 0 0 8 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-.6-1 1.7 1.7 0 0 0-1.2-.5H3a2 2 0 0 1 0-4h.09A1.7 1.7 0 0 0 4.6 8a1.7 1.7 0 0 0-.34-1.88l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-.6 1.7 1.7 0 0 0 .5-1.2V3a2 2 0 0 1 4 0v.09A1.7 1.7 0 0 0 16 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.7 1.7 0 0 0 19.4 9c.4.2.7.5 1 .6.4.2.8.2 1.2.2H21a2 2 0 0 1 0 4h-.09A1.7 1.7 0 0 0 19.4 15z"></path>
            </svg>
          </button>
          <button class="win10-start-side-btn" :title="$t('win10.startMenu.actions.power')" @click="shutdown">
            <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" height="16" width="16">
              <path d="M12 2v10"></path>
              <path d="M18.4 6.6a9 9 0 1 1-12.8 0"></path>
            </svg>
          </button>
        </div>
      </div>
      <div class="w-64 py-2 overflow-y-hidden hover:overflow-y-auto">
        <div class="px-3 pb-2 text-xs text-gray-200">{{ $t("win10.startMenu.sections.allApps") }}</div>
        <div v-for="app in visibleApps" :key="app.code">
          <button
            class="win10-start-app"
            @click="openApp(app)"
            v-context-menu="appContext(app)"
          >
            <IconView :icon="app.icon" :alt="appName(app)" :size="24" icon-class="w-6 h-6 object-contain" />
            <span class="text-xs truncate">{{ appName(app) }}</span>
          </button>
        </div>
      </div>
      <div class="flex-1 py-3 px-2 overflow-y-auto">
        <div class="text-xs mb-2 text-gray-200">{{ $t("win10.startMenu.sections.myApps") }}</div>
        <div class="grid grid-cols-2 gap-1">
          <button
            v-for="app in pinnedApps"
            :key="app.code"
            class="win10-start-tile"
            @click="openApp(app)"
            v-context-menu="pinnedContext(app)"
          >
            <IconView :icon="app.icon" :alt="appName(app)" :size="40" icon-class="w-10 h-10 object-contain mx-auto" />
            <span class="text-xs truncate w-full text-center">{{ appName(app) }}</span>
          </button>
          <div v-if="pinnedApps.length === 0" class="col-span-2 text-xs text-gray-300/70 px-1 py-3">
            {{ $t("win10.startMenu.tips.pinToMyApps") }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
