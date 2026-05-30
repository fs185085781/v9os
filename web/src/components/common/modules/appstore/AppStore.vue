<script setup>
import { computed, onMounted, ref } from "vue";
import {
  NBadge,
  NButton,
  NCard,
  NEmpty,
  NInput,
  NLayout,
  NLayoutSider,
  NMenu,
  NScrollbar,
  NSpin,
  NTag,
} from "naive-ui";
import { checkAuth } from "@/directives/auth";
import { useStore } from "@/stores/user.js";
import { useEventBus } from "@/util/event.js";
import { getWinSize, postData, postStreamData,absoluteUrl } from "@/util/util";
import AppStoreAdd from "./AppStoreAdd.vue";
import AppStoreDetail from "./AppStoreDetail.vue";

const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});

const user = useStore();

const categories = ref([]);
const apps = ref([]);
const installedApps = ref([]);
const loading = ref(false);
const installStates = ref({});
const searchKeyword = ref("");
const activeMenu = ref("all");
const siderCollapsed = ref(false);

const canInstall = computed(() => checkAuth("/api/appstore/install"));
const canAdd = computed(() => checkAuth("/api/appstore/add"));
const canViewInstalled = computed(() => checkAuth("/api/appstore/installed"));
const canUpgrade = computed(() => checkAuth("/api/appstore/install"));

const getString = (record, ...keys) => {
  for (const key of keys) {
    const value = record?.[key];
    if (value !== undefined && value !== null && String(value).trim() !== "") {
      return String(value).trim();
    }
  }
  return "";
};

const getNumber = (record, ...keys) => {
  for (const key of keys) {
    const value = record?.[key];
    if (value !== undefined && value !== null && value !== "") {
      const num = Number(value);
      if (!Number.isNaN(num)) {
        return num;
      }
    }
  }
  return 0;
};

const getBool = (record, ...keys) => {
  for (const key of keys) {
    const value = record?.[key];
    if (value !== undefined && value !== null) {
      return value === true || value === "true" || value === 1 || value === "1";
    }
  }
  return false;
};

const normalizeScreenshots = (record) => {
  const list = record?.Screenshots ?? record?.screenshots ?? [];
  return Array.isArray(list) ? list.filter(Boolean) : [];
};

const normalizeInstallable = (record) => {
  if (record?.Installable === undefined && record?.installable === undefined) {
    return true;
  }
  return getBool(record, "Installable", "installable");
};

const normalizePackages = (record) => {
  const list = record?.Packages ?? record?.packages ?? [];
  if (!Array.isArray(list)) {
    return [];
  }
  return list.map((item) => ({
    OS: getString(item, "OS", "os"),
    Arch: getString(item, "Arch", "arch"),
    PackageName: getString(item, "PackageName", "packageName"),
    PackageHash: getString(item, "PackageHash", "packageHash"),
    DownloadUrl: getString(item, "DownloadURL", "DownloadUrl", "downloadUrl"),
    PackageSize: getNumber(item, "PackageSize", "packageSize"),
  }));
};

const normalizeRuntime = (record = {}) => {
  const runtime = record?.Runtime ?? record?.runtime ?? {};
  return {
    AccessUrl: getString(runtime, "AccessUrl", "accessUrl") || getString(record, "AccessUrl", "accessUrl"),
  };
};

const normalizeApp = (record = {}) => ({
  ...record,
  Code: getString(record, "Code", "code"),
  Name: getString(record, "Name", "name"),
  Description: getString(record, "Description", "description"),
  IconUrl: getString(record, "IconUrl", "iconUrl"),
  Author: getString(record, "Author", "author"),
  Version: getString(record, "Version", "version"),
  InstalledVersion: getString(record, "InstalledVersion", "installedVersion"),
  Category: getString(record, "Category", "category"),
  PluginType: getNumber(record, "PluginType", "pluginType"),
  LimitVersion: getString(record, "LimitVersion", "limitVersion"),
  Installable: normalizeInstallable(record),
  InstallReason: getString(record, "InstallReason", "installReason"),
  Installed: getBool(record, "Installed", "installed"),
  Runtime: normalizeRuntime(record),
  Screenshots: normalizeScreenshots(record),
  Packages: normalizePackages(record),
});

const extractAppList = (payload) => {
  if (Array.isArray(payload)) {
    return payload;
  }
  if (Array.isArray(payload?.Data)) {
    return payload.Data;
  }
  if (Array.isArray(payload?.data)) {
    return payload.data;
  }
  if (Array.isArray(payload?.List)) {
    return payload.List;
  }
  if (Array.isArray(payload?.list)) {
    return payload.list;
  }
  return [];
};

const compareVersions = (a, b) => {
  const aList = String(a || "")
    .split(".")
    .map((item) => Number(item) || 0);
  const bList = String(b || "")
    .split(".")
    .map((item) => Number(item) || 0);
  const length = Math.max(aList.length, bList.length);
  for (let i = 0; i < length; i += 1) {
    const left = aList[i] || 0;
    const right = bList[i] || 0;
    if (left > right) {
      return 1;
    }
    if (left < right) {
      return -1;
    }
  }
  return 0;
};

const pluginTypeText = (type) => {
  switch (type) {
    case 1:
      return $t("common.appstore.pluginTypes.main");
    case 2:
      return $t("common.appstore.pluginTypes.frontend");
    case 3:
      return $t("common.appstore.pluginTypes.thirdParty");
    case 4:
      return $t("common.appstore.pluginTypes.remote");
    default:
      return $t("common.appstore.pluginTypes.unknown");
  }
};

const isInstallable = (app) => app?.Installable !== false;

const getInstallState = (code) => installStates.value[code] || null;

const isInstalling = (code) => !!getInstallState(code);

const setInstallState = (code, state) => {
  installStates.value = {
    ...installStates.value,
    [code]: state,
  };
};

const clearInstallState = (code) => {
  const next = { ...installStates.value };
  delete next[code];
  installStates.value = next;
};

const installProgressText = (code) => {
  const state = getInstallState(code);
  if (!state) {
    return "";
  }
  const progress = Number(state.progress) || 0;
  return `${progress}%`;
};

const menuOptions = computed(() => {
  const options = [{ label: $t("common.appstore.all"), key: "all" }];
  if (canViewInstalled.value) {
    options.push({ label: $t("common.appstore.installed"), key: "installed" });
  }
  options.push({ type: "divider", key: "divider-1" });
  categories.value.forEach((item) => {
    const name = item.name || item.Name;
    if (!name) {
      return;
    }
    options.push({
      label: name,
      key: `cat_${encodeURIComponent(name)}`,
    });
  });
  return options;
});

const isInstalled = (code) =>
  installedApps.value.some((item) => item.Code === code);

const getInstalledApp = (code) =>
  installedApps.value.find((item) => item.Code === code) || null;

const hasUpgrade = (app) => {
  const installed = getInstalledApp(app.Code);
  if (!installed || !app.Version) {
    return false;
  }
  return compareVersions(app.Version, installed.Version) > 0;
};

const loadCategories = async () => {
  const data = await postData("appstore", "categories", {});
  categories.value = Array.isArray(data) ? data : [];
};

const loadApps = async (category = "") => {
  loading.value = true;
  const data = await postData("appstore", "apps", {
    category,
    page: 1,
    pageSize: 50,
  });
  apps.value = extractAppList(data).map(normalizeApp);
  appInitIcons(apps.value)
  loading.value = false;
};

const loadInstalled = async () => {
  if (!canViewInstalled.value) {
    installedApps.value = [];
    return;
  }
  const data = await postData("appstore", "installed", {});
  installedApps.value = extractAppList(data).map((item) => {
    if (item.IconUrl && item.IconUrl.startsWith("/")) {
      item.IconUrl = absoluteUrl(item.IconUrl);
    }
    return normalizeApp({
      ...item,
      Installed: true,
    });
  });
};

const loadInstalledApps = async () => {
  loading.value = true;
  await loadInstalled();
  apps.value = installedApps.value;
  appInitIcons(apps.value)
  loading.value = false;
};

const searchApps = async () => {
  if (!searchKeyword.value.trim()) {
    await handleMenuSelect(activeMenu.value);
    return;
  }
  loading.value = true;
  const data = await postData("appstore", "search", {
    keyword: searchKeyword.value.trim(),
    page: 1,
    pageSize: 50,
  });
  apps.value = extractAppList(data).map(normalizeApp);
  appInitIcons(apps.value)
  loading.value = false;
};

const appInitIcons = (list) => {
  list.forEach(item => {

  })
}

const refreshCurrentView = async () => {
  if (searchKeyword.value.trim()) {
    await searchApps();
    return;
  }
  if (activeMenu.value === "installed") {
    await loadInstalledApps();
    return;
  }
  if (activeMenu.value === "all") {
    await loadApps("");
    return;
  }
  if (activeMenu.value.startsWith("cat_")) {
    await loadApps(decodeURIComponent(activeMenu.value.replace("cat_", "")));
  }
};

const refreshAfterMutation = async () => {
  await loadInstalled();
  await refreshCurrentView();
  await $webhook.refresh();
  await user.loadUser();
};

const installOrUpgradeApp = async (type, code) => {
  if (isInstalling(code)) {
    return;
  }
  setInstallState(code, { progress: 1 });
  let url = new URL(absoluteUrl());
  const hostUrl = url.protocol + "//" + url.hostname;
  let success = false;
  let failed = false;
  try {
    const result = await postStreamData(
      "appstore",
      "install",
      {
        type,
        code,
        hostUrl
      },
      (data) => {
        if (data?.code === -1) {
          failed = true;
          $msg.message.error(data.msg);
          return;
        }
        if (data?.progress !== undefined) {
          const progress = Math.max(1, Math.min(100, Number(data.progress) || 1));
          setInstallState(code, { progress });
        }
        if (data?.done === true && data?.code === 0) {
          success = true;
        }
      },
      "okerr",
    );
    if (result && success && !failed) {
      await refreshAfterMutation();
    }
  } catch (err) {
    $msg.message.error(err?.message || String(err));
  } finally {
    clearInstallState(code);
  }
};

const installApp = async (code) => {
  await installOrUpgradeApp("1", code);
};

const upgradeApp = async (code) => {
  await installOrUpgradeApp("2", code);
};

const handleMenuSelect = async (key) => {
  activeMenu.value = key;
  searchKeyword.value = "";
  if (key === "installed") {
    await loadInstalledApps();
    return;
  }
  if (key === "all") {
    await loadApps("");
    return;
  }
  if (key.startsWith("cat_")) {
    await loadApps(decodeURIComponent(key.replace("cat_", "")));
  }
};

const openDetail = (app) => {
  const size = getWinSize(900);
  const item = normalizeApp(app);
  $wins.addWindow(
    {
      width: size.width,
      height: size.height,
      title: item.Name || item.Code || $t("common.appstore.detailTitle"),
      component: AppStoreDetail,
      data: {
        code: item.Code,
      },
    },
    props.winId,
  );
};

const openAddApp = () => {
  const size = getWinSize(720);
  $wins.addWindow(
    {
      width: size.width,
      height: size.height,
      title: $t("common.appstore.addApp"),
      component: AppStoreAdd,
    },
    props.winId,
  );
};

const openPlugin = async (app) => {
  const item = normalizeApp(app);
  let path = `/page/${item.Code}/`;
  if (item.PluginType === 2) {
    path = `/api/webplugin/${item.Code}/`;
  } else if (item.PluginType === 3) {
    path = `/api/thirdplugin/${item.Code}/`;
  } else if (item.PluginType === 4) {
    path = item.Runtime?.AccessUrl || getInstalledApp(item.Code)?.Runtime?.AccessUrl;
    if (!path) {
      $msg.message.error($t("common.appstore.openFailed"));
      return;
    }
  }
  const size = getWinSize();
  $wins.addWindow({
    icon:app.IconUrl,
    width: size.width,
    height: size.height,
    title: item.Name || item.Code,
    iframeUrl: absoluteUrl(`${path}`),
  });
};

useEventBus("appstore-refresh", async () => {
  await refreshAfterMutation();
});

onMounted(async () => {
  const tasks = [loadCategories()];
  if (canViewInstalled.value) {
    tasks.push(loadInstalled());
  }
  await Promise.all(tasks);
  await loadApps("");
});
</script>

<template>
  <div class="w-full h-full overflow-hidden">
    <n-layout has-sider class="h-full">
      <n-layout-sider :collapsed="siderCollapsed" :collapsed-width="0" :width="240" collapse-mode="width"
        show-trigger="bar" :native-scrollbar="false" class="h-full border-r border-[var(--user-divider-color)]" @collapse="siderCollapsed = true"
        @expand="siderCollapsed = false">
        <div class="px-5 pt-4 pb-3 text-5 font-700">{{ $t("common.appstore.title") }}</div>
        <n-menu :options="menuOptions" :value="activeMenu" @update:value="handleMenuSelect" />
      </n-layout-sider>

      <n-layout class="h-full">
        <div class="flex flex-col h-full">
          <div class="flex gap-2 px-4 py-3">
            <n-input v-model:value="searchKeyword" :placeholder="$t('common.appstore.searchPlaceholder')" clearable
              @keyup.enter="searchApps" @clear="handleMenuSelect(activeMenu)">
              <template #suffix>
                <n-button text size="small" :disabled="!searchKeyword.trim()" @click="searchApps">
                  {{ $t("common.appstore.search") }}
                </n-button>
              </template>
            </n-input>
            <n-button v-if="canAdd" type="primary" @click="openAddApp">
              {{ $t("common.appstore.addApp") }}
            </n-button>
          </div>

          <n-scrollbar class="flex-1">
            <n-spin :show="loading" class="min-h-60">
              <div v-if="!loading && apps.length === 0" class="min-h-60 flex items-center justify-center">
                <n-empty :description="$t('common.appstore.empty')" />
              </div>

              <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3 pt-1 px-4 pb-4">
                <div v-for="app in apps" :key="app.Code" class="cursor-pointer" @click="openDetail(app)">
                  <n-card hoverable class="transition-transform duration-200 hover:-translate-y-0.5">
                    <div class="grid grid-cols-[64px_minmax(0,1fr)] gap-3">
                      <div class="relative shrink-0">
                        <img v-if="app.IconUrl" :src="absoluteUrl(app.IconUrl)" class="w-16 h-16 object-cover user-rounded-4" />
                        <div v-else class="w-16 h-16 flex items-center justify-center user-rounded-4 user-color-surface-muted">?</div>
                        <n-badge v-if="hasUpgrade(app)" dot :offset="[-4, 4]" />
                      </div>

                      <div class="flex-1 min-w-0">
                        <div class="text-3.75 font-600 leading-1.3">{{ app.Name || app.Code }}</div>
                        <div class="mt-1.25 text-3 user-color-muted">
                          {{ app.Author || pluginTypeText(app.PluginType) }}
                        </div>
                        <div class="mt-0 max-w-full h-[34px] text-3 leading-[17px] opacity-75 overflow-hidden text-ellipsis [display:-webkit-box] [-webkit-line-clamp:2] [-webkit-box-orient:vertical]">
                          {{
                            app.Description ||
                            $t("common.appstore.noDescription")
                          }}
                        </div>
                        <div class="flex gap-1.5 mt-2 flex-wrap">
                          <n-tag size="small" round>{{
                            pluginTypeText(app.PluginType)
                          }}</n-tag>
                          <n-tag v-if="isInstalled(app.Code)" type="success" size="small" round>{{
                            $t("common.appstore.installed")
                          }}</n-tag>
                        </div>
                      </div>

                      <div class="col-span-2 flex justify-end">
                        <n-button v-if="canInstall && !isInstalled(app.Code)" type="primary" size="small" round
                          :loading="isInstalling(app.Code)" :disabled="!isInstallable(app) || isInstalling(app.Code)"
                          @click.stop="installApp(app.Code)">
                          {{
                            isInstalling(app.Code)
                              ? installProgressText(app.Code)
                              : isInstallable(app)
                                ? $t("common.appstore.install")
                                : $t("common.appstore.incompatible")
                          }}
                        </n-button>
                        <n-button v-else-if="canUpgrade && hasUpgrade(app)" type="info" size="small" round
                          :loading="isInstalling(app.Code)" :disabled="isInstalling(app.Code)"
                          @click.stop="upgradeApp(app.Code)">
                          {{ isInstalling(app.Code) ? installProgressText(app.Code) : $t("common.appstore.upgrade") }}
                        </n-button>
                        <n-button v-else-if="isInstalled(app.Code)" size="small" round @click.stop="openPlugin(app)">
                          {{ $t("common.appstore.open") }}
                        </n-button>
                      </div>
                    </div>
                  </n-card>
                </div>
              </div>
            </n-spin>
          </n-scrollbar>
        </div>
      </n-layout>
    </n-layout>
  </div>
</template>
