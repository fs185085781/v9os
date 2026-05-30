<script setup>
import {
  h,
  ref,
  reactive,
  computed,
  defineAsyncComponent,
  watch,
  onMounted,
  onBeforeUnmount,
} from "vue";
import {
  NLayout,
  NLayoutSider,
  NLayoutHeader,
  NLayoutContent,
  NMenu,
  NBreadcrumb,
  NBreadcrumbItem,
  NTabs,
  NTabPane,
  NDropdown,
  NIcon,
} from "naive-ui";
import {
  FullscreenOutlined,
  FullscreenExitOutlined,
  SettingOutlined,
  MessageOutlined,
  DownOutlined,
} from "@vicons/antd";
import { checkAuth } from "@/directives/auth";
import { useEventBus } from "@/util/event.js";
import { getWinSize } from "@/util/util.js";
import ChatNotifyPopover from "@/components/common/component/user/chat/ChatNotifyPopover.vue";
import { adminMenu } from "@/components/common/views/admin.js";
import { AppFolder24Regular, Apps24Regular } from "@vicons/fluent";
import { renderIcon } from "@/util/icon"
const headerIconSize = 28;
const settingsHeaderIcon = renderIcon(SettingOutlined, headerIconSize);
const notificationHeaderIcon = renderIcon(MessageOutlined, headerIconSize);
const pageMap = {};
const menuData = adminMenu()
const pluginDataMap = ref({})
const appInfoMaps = ref({});
const dockAppKeys = ref([]);
const dynamicMenuData = ref([]);
const showMenuData = computed(() => [...menuData, ...dynamicMenuData.value]);
const initMenus = async () => {
  const apps = await $user.getMyApps();
  const dockApps = await $user.getDockApps();
  const allChildren = [];
  dockAppKeys.value = [];
  apps.forEach((app) => {
    if (app.code == "__kernel__") {
      return;
    }
    const key = "plugin." + app.code;
    appInfoMaps.value[key] = app;
    const favorite = dockApps.includes(app.code);
    if (favorite) {
      dockAppKeys.value.push(key);
    }
    allChildren.push({
      key,
      favorite,
      icon: renderIcon(Apps24Regular),
    })
    pluginDataMap.value["plugin." + app.code] = app;
  });
  allChildren.sort((a, b) => b.favorite - a.favorite);
  dynamicMenuData.value = [];
  dynamicMenuData.value.push({
    key: "allapp",
    icon: renderIcon(AppFolder24Regular),
    children: allChildren
  })
  for (const value of showMenuData.value) {
    for (const child of value.children) {
      let com = null;
      if (child.key.startsWith("inface.") || child.key.startsWith("component.")) {
        const modsMap = child.key.startsWith("inface.") ? window.__INFACE_MODS__ : window.__COMPONENT_MODS__;
        const mod = modsMap[`/src/components/common/${child.key.replaceAll(".", "/")}.vue`]
        if (!mod) {
          continue;
        }
        com = defineAsyncComponent(
          mod
        );
      } else if (child.key.startsWith("plugin.")) {
        com = defineAsyncComponent(
          () =>
            import(
              `./PluginPanel.vue`
            ),
        );
      } else {
        com = defineAsyncComponent(
          () =>
            import(
              `@/components/common/views/${child.pkey}/${child.key}/index.vue`
            ),
        );
      }
      pageMap[child.key] = com;
    }
  }
}
function filterMenuByAuth() {
  const tables = [];
  for (const value of showMenuData.value) {
    const table = {
      key: value.key,
      icon: value.icon,
      children: [],
    };
    table.label = $t("admin.system." + table.key);
    for (const child of value.children) {
      if (!pageMap[child.key]) {
        continue;
      }
      if (!checkAuth(child.auth || `/api/${child.key}/page`) && !appInfoMaps.value[child.key]) {
        continue;
      }
      table.children.push({
        key: child.key,
        icon: child.icon,
        label: getMenuLabel(child.key, true),
      });
    }
    if (table.children.length > 0) {
      tables.push(table);
    }
  }
  return tables;
}

// 4. 用户数据
const userData = computed(() => {
  return $user.user;
});

// 5. 面包屑数据
const breadcrumbs = ref([]);
// 6. 标签页数据
const tabs = ref([]);
const activeTab = ref("");
const openProfileTab = () => {
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width < 740 ? sz.width : 740,
    height: sz.height,
    title: $tc("backend.layout.profile"),
    component: defineAsyncComponent(
      () => import("@/components/common/component/user/profile/profile.vue"),
    ),
    data: {},
  });
};

const openPersonalizeTab = () => {
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width < 740 ? sz.width : 740,
    height: sz.height,
    title: $tc("backend.layout.settings"),
    component: defineAsyncComponent(
      () => import("@/components/common/modules/settings/PersonalizeTab.vue"),
    ),
    data: {},
  });
};

// 监听 activeTab 变化，更新面包屑
watch(activeTab, (newKey) => {
  if (newKey) {
    updateBreadcrumbs(newKey);
  }
});
const visibleTabs = computed(() => tabs.value.filter((tab) => pageMap[tab]));
const tabMenuState = reactive({
  show: false,
  x: 0,
  y: 0,
  tabKey: "",
});

const tabMenuOptions = computed(() => {
  const s1 = [
    {
      label: "关闭当前",
      key: "close-current",
      disabled: !tabMenuState.tabKey,
    },
    {
      label: "关闭其他",
      key: "close-others",
      disabled: tabs.value.length <= 1,
    },
    {
      label: "关闭右侧",
      key: "close-right",
      disabled:
        !tabMenuState.tabKey ||
        tabs.value.findIndex((tab) => tab === tabMenuState.tabKey) === -1 ||
        tabs.value.findIndex((tab) => tab === tabMenuState.tabKey) ===
        tabs.value.length - 1,
    },
    {
      label: "全部关闭",
      key: "close-all",
      disabled: tabs.value.length === 0,
    }
  ];
  const s2 = [];
  if (dockAppKeys.value.includes(tabMenuState.tabKey)) {
    s2.push({
      label: "取消收藏",
      key: "cancel-favorite",
    });
  } else if (appInfoMaps.value[tabMenuState.tabKey]) {
    s2.push({
      label: "收藏",
      key: "favorite",
    });
  }
  return [...s1, ...s2];
});

// 添加图标到菜单数据（带权限过滤）
const menuOptions = computed(() => {
  return filterMenuByAuth();
});

// 获取菜单项的翻译标题
function getMenuLabel(key, isLeft) {
  let label = null;
  if (key.startsWith("inface.")) {
    label = $t(key.replace("ee.", "ee.tabs."));
  }else if(key.startsWith("component.")){
    label = $t(key.replace("component.", "component.tabs."));
  } else if (key.startsWith("plugin.")) {
    const n = appInfoMaps.value[key].name;
    const favorite = dockAppKeys.value.includes(key);
    label = n && n.value ? n.value : n;
    if (isLeft && favorite) {
      label = "❤️" + label;
    }
  } else {
    label = $t(`model.${key}.model`) + $t(`admin.table.model`);
  }
  return label;
}

// 获取父级菜单的翻译标题
function getParentLabel(key, level) {
  return level == 0 ? $t("admin.system." + key) : getMenuLabel(key, false);
}

// 查找菜单路径（返回带 key 的路径）
function findMenuPath(items, key, path = [], level) {
  for (const item of items) {
    const currentPath = [
      ...path,
      {
        key: item.key,
        level: level,
      },
    ];
    if (item.key === key) return currentPath;
    if (item.children) {
      const found = findMenuPath(item.children, key, currentPath, level + 1);
      if (found) return found;
    }
  }
  return null;
}

// ==================== 核心联动逻辑 ====================

// 1. 点击左侧菜单 → 添加标签 + 更新面包屑 + 加载组件
function handleMenuSelect(key) {
  // 添加标签（如果不存在）
  const existTab = tabs.value.find((t) => t === key);
  if (!existTab) {
    tabs.value.push(key);
  }
  // 切换到该标签
  activeTab.value = key;
  // 更新面包屑
  updateBreadcrumbs(key);
}

// 2. 关闭标签
function handleTabClose(key) {
  const index = tabs.value.findIndex((t) => t === key);
  tabs.value = tabs.value.filter((t) => t !== key);
  // 如果关闭的是当前激活标签，切换到最后一个
  if (activeTab.value === key && tabs.value.length > 0) {
    activeTab.value = tabs.value[Math.min(index, tabs.value.length - 1)];
    updateBreadcrumbs(activeTab.value);
  } else if (tabs.value.length === 0) {
    activeTab.value = "";
    breadcrumbs.value = [];
  }
}

function closeCurrentTab() {
  if (!tabMenuState.tabKey) {
    return;
  }
  handleTabClose(tabMenuState.tabKey);
}

function closeRightTabs() {
  if (!tabMenuState.tabKey) {
    return;
  }
  const index = tabs.value.findIndex((tab) => tab === tabMenuState.tabKey);
  if (index === -1 || index === tabs.value.length - 1) {
    return;
  }
  const rightTabs = tabs.value.slice(index + 1);
  tabs.value = tabs.value.slice(0, index + 1);
  if (rightTabs.includes(activeTab.value)) {
    activeTab.value = tabMenuState.tabKey;
    updateBreadcrumbs(activeTab.value);
  }
}

function closeAllTabs() {
  tabs.value = [];
  activeTab.value = "";
  breadcrumbs.value = [];
}

function closeOtherTabs() {
  if (!tabMenuState.tabKey) {
    return;
  }
  // 只保留当前标签
  tabs.value = [tabMenuState.tabKey];
  activeTab.value = tabMenuState.tabKey;
  updateBreadcrumbs(activeTab.value);
}

function showTabMenu(tab, event) {
  event.preventDefault();
  event.stopPropagation();
  tabMenuState.tabKey = tab;
  tabMenuState.x = event.clientX;
  tabMenuState.y = event.clientY;
  tabMenuState.show = true;
}

function hideTabMenu() {
  tabMenuState.show = false;
}

function handleTabMenuSelect(key) {
  if (key === "close-current") {
    closeCurrentTab();
  } else if (key === "close-others") {
    closeOtherTabs();
  } else if (key === "close-right") {
    closeRightTabs();
  } else if (key === "close-all") {
    closeAllTabs();
  } else if (key === "cancel-favorite") {
    $user.removeDockApp(tabMenuState.tabKey.split(".")[1]);
  } else if (key === "favorite") {
    $user.addDockApp(tabMenuState.tabKey.split(".")[1]);
  }
  hideTabMenu();
}

function renderTabLabel(tab) {
  return h(
    "div",
    {
      class: "tab-label",
      onContextmenu: (event) => showTabMenu(tab, event),
    },
    getMenuLabel(tab, false),
  );
}

// 3. 更新面包屑
function updateBreadcrumbs(key) {
  const path = findMenuPath(showMenuData.value, key, [], 0);
  if (path) {
    breadcrumbs.value = path;
  }
}

// 4. 用户下拉菜单选择
function avatarSelect(key) {
  if (key === "profile") openProfileTab();
  else if (key === "settings") openPersonalizeTab();
  else if (key === "logout") $user.logout();
}

// 5. 用户下拉菜单选项
const avatarOptions = computed(() => {
  return [
    { label: $t("backend.layout.profile"), key: "profile" },
    { label: $t("backend.layout.settings"), key: "settings" },
    { label: $t("backend.layout.logout"), key: "logout" },
  ];
});

// 6. 侧边栏折叠状态
const collapsed = ref(false);

// 7. 全屏状态
const isFullscreen = ref(false);
const fullscreenHeaderIcon = computed(() => renderIcon(isFullscreen.value ? FullscreenExitOutlined : FullscreenOutlined, headerIconSize));

function toggleFullscreen() {
  if (!isFullscreen.value) {
    document.documentElement.requestFullscreen();
  } else {
    document.exitFullscreen();
  }
}

// 监听全屏状态变化
function handleFullscreenChange() {
  const fele = document.fullscreenElement || document.mozFullScreenElement || document.webkitFullscreenElement;
  isFullscreen.value = !!fele;
}
// 9. 初始化：自动选中第一个有权限的菜单项
function initDefaultMenu() {
  const filtered = filterMenuByAuth();
  if (
    filtered.length > 0 &&
    filtered[0].children &&
    filtered[0].children.length > 0
  ) {
    const firstKey = filtered[0].children[0].key;
    activeTab.value = firstKey;
    tabs.value.push(firstKey);
    updateBreadcrumbs(firstKey);
  }
}
function openSettings() {
  const sz = getWinSize();
  $wins.addWindow({
    width: sz.width,
    height: sz.height,
    title: $tc("common.settings.title"),
    icon: "/assets/deepin/img/icons/settings.png",
    component: defineAsyncComponent(
      () => import("./SystemPreferences.vue"),
    ),
    data: {},
  });
}

onMounted(async () => {
  window.addEventListener("click", hideTabMenu);
  document.addEventListener('fullscreenchange', handleFullscreenChange);
  document.addEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.addEventListener('mozfullscreenchange', handleFullscreenChange);

  await initMenus()
  initDefaultMenu();
});

onBeforeUnmount(() => {
  window.removeEventListener("click", hideTabMenu);
  document.removeEventListener('fullscreenchange', handleFullscreenChange);
  document.removeEventListener('webkitfullscreenchange', handleFullscreenChange);
  document.removeEventListener('mozfullscreenchange', handleFullscreenChange);
});

useEventBus("dock-apps-change", () => {
  initMenus();
});
</script>

<template>
  <n-layout class="h-screen w-screen" has-sider>
    <!-- ==================== 左侧边栏 ==================== -->
    <n-layout-sider collapse-mode="width" :collapsed-width="0" :width="220" :collapsed="collapsed" show-trigger
      @collapse="collapsed = true" @expand="collapsed = false"
      class="user-color-fbg user-color-ftext border-r border-gray-200">
      <!-- Logo -->
      <div class="h-64px flex items-center justify-center border-b border-gray-200">
        <div class="flex items-center gap-10px" v-show="!collapsed">
          <div class="w-32px h-32px bg-blue-500 rounded flex items-center justify-center">
            <img :src="$user.webSettings.Logo" alt="Logo" class="w-32px h-32px" />
          </div>
          <span class="text-16px font-bold user-color-text">{{
            $user.webSettings.Title
            }}</span>
        </div>
      </div>

      <!-- 菜单 -->
      <div class="h-[calc(100vh-64px)] overflow-y-auto">
        <n-menu :options="menuOptions" :collapsed="collapsed" :collapsed-icon-size="28"
          :default-expanded-keys="['system', 'list', 'form']" :value="activeTab" @update:value="handleMenuSelect" />
      </div>
    </n-layout-sider>

    <!-- ==================== 右侧内容区 ==================== -->
    <n-layout>
      <!-- 顶部 Header -->
      <n-layout-header
        class="h-64px user-color-fbg user-color-ftext flex items-center justify-between px-16px border-b border-gray-200">
        <!-- 左侧：面包屑 -->
        <div class="flex items-center gap-12px">
          <n-breadcrumb separator="/">
            <n-breadcrumb-item v-for="item in breadcrumbs" :key="item.key">
              {{ getParentLabel(item.key, item.level) }}
            </n-breadcrumb-item>
          </n-breadcrumb>
        </div>

        <!-- 右侧：功能按钮 -->
        <div class="flex items-center gap-8px">
          <!-- 全屏 -->
          <div
            class="w-36px h-36px flex items-center justify-center cursor-pointer hover:user-color-bg rounded transition-colors"
            @click="toggleFullscreen">
            <component :is="fullscreenHeaderIcon" />
          </div>

          <ChatNotifyPopover placement="bottom-end" :icon-size="18"
            :icon-render="notificationHeaderIcon"
            trigger-class="w-36px h-36px flex items-center justify-center cursor-pointer hover:user-color-bg rounded transition-colors border-0 bg-transparent user-color-ftext" />

          <!-- 设置 -->
          <div
            class="w-36px h-36px flex items-center justify-center cursor-pointer hover:user-color-bg rounded transition-colors"
            @click="openSettings">
            <component :is="settingsHeaderIcon" />
          </div>

          <!-- 用户头像 -->
          <n-dropdown trigger="hover" :options="avatarOptions" @select="avatarSelect">
            <div
              class="flex items-center gap-8px cursor-pointer px-12px py-6px hover:bg-gray-100 rounded transition-colors">
              <img v-if="userData.Avatar" :src="userData.Avatar"
                class="w-8 h-8 rounded-full mr-4 shrink-0 object-cover" />
              <div v-else
                class="w-8 h-8 rounded-full mr-4 shrink-0 flex items-center justify-center user-color-bg text-2xl font-600">
                {{ userData.Username?.charAt(0)?.toUpperCase() }}
              </div>
              <span class="text-13px text-gray-600">{{ userData.Name }}</span>
              <n-icon size="12">
                <DownOutlined />
              </n-icon>
            </div>
          </n-dropdown>
        </div>
      </n-layout-header>

      <!-- 多标签栏 -->
      <div class="user-color-fbg user-color-ftext border-b border-gray-200 px-12px pt-8px">
        <n-tabs v-model:value="activeTab" type="card" size="medium" closable @close="handleTabClose">
          <n-tab-pane v-for="tab in tabs" :key="tab" :name="tab" :tab="renderTabLabel(tab)" />
        </n-tabs>
      </div>

      <!-- 内容区域 -->
      <n-layout-content class="user-color-fbg user-color-ftext overflow-y-auto" style="height: calc(100vh - 130px)">
        <div v-for="tab in visibleTabs" :key="tab" v-show="activeTab === tab" class="h-full">
          <component :is="pageMap[tab]" :data="pluginDataMap[tab]" />
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>

  <n-dropdown trigger="manual" placement="bottom-start" :show="tabMenuState.show" :options="tabMenuOptions"
    :x="tabMenuState.x" :y="tabMenuState.y" @clickoutside="hideTabMenu" @select="handleTabMenuSelect" />
</template>

<style scoped>
.tab-label {
  user-select: none;
}
</style>
