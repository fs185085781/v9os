<script setup>
import { computed, ref, watch } from "vue";
import { NButton, NEmpty, NImage, NSpace, NSpin } from "naive-ui";
import { checkAuth } from "@/directives/auth";
import { webhookStore } from "@/stores/webhook.js";
import { useStore } from "@/stores/user.js";
import emitter from "@/util/event.js";
import { getApiHost, getWinSize, postData, postStreamData } from "@/util/util";

const props = defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
  winId: {
    type: String,
    default: "",
  },
});
const user = useStore();
const hooks = webhookStore();

const selectedApp = ref(null);
const appVersions = ref([]);
const detailLoading = ref(false);
const installState = ref(null);
const screenshotIndex = ref(0);

const canInstall = computed(() => checkAuth("/api/appstore/install"));
const canUninstall = computed(() => checkAuth("/api/appstore/uninstall"));
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

const normalizeApp = (record = {}) => ({
  ...record,
  Code: getString(record, "Code", "code"),
  Name: getString(record, "Name", "name"),
  Description: getString(record, "Description", "description"),
  IconUrl: getString(record, "IconUrl", "iconUrl"),
  Author: getString(record, "Author", "author"),
  Remark: getString(record, "Remark", "remark"),
  Version: getString(record, "Version", "version"),
  StoreVersion: getString(record, "StoreVersion", "storeVersion"),
  Category: getString(record, "Category", "category"),
  PluginType: getNumber(record, "PluginType", "pluginType"),
  AccessUrl: getString(record, "AccessUrl", "accessUrl"),
  RuntimeError: getString(record, "RuntimeError", "runtimeError"),
  FirstMachine: getString(record, "FirstMachine", "firstMachine"),
  CloseDelay: getNumber(record, "CloseDelay", "closeDelay"),
  Status: getNumber(record, "Status", "status"),
  NeedLogin: getNumber(record, "NeedLogin", "needLogin"),
  Interceptors: getString(record, "Interceptors", "interceptors"),
  WebHook: getString(record, "WebHook", "webHook"),
  LimitVersion: getString(record, "LimitVersion", "limitVersion"),
  DebugPort: getNumber(record, "DebugPort", "debugPort"),
  Installable: normalizeInstallable(record),
  InstallReason: getString(record, "InstallReason", "installReason"),
  Installed: getBool(record, "Installed", "installed"),
  Screenshots: normalizeScreenshots(record),
  Packages: normalizePackages(record),
});

const normalizeVersion = (record = {}) => ({
  Version: getString(record, "Version", "version"),
  Changelog: getString(
    record,
    "Changelog",
    "changelog",
    "Notes",
    "notes",
    "Description",
    "description",
  ),
  CreatedAt: getString(record, "CreatedAt", "createdAt", "Date", "date"),
  Packages: normalizePackages(record),
});

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

const packageTargetText = (pkg) => {
  const os = String(pkg?.OS || "");
  const arch = String(pkg?.Arch || "");
  if (os === "all" && arch === "all") {
    return $t("common.appstore.universalPackage");
  }
  return `${os}/${arch}`;
};

const versionPackageSummary = (packages) => {
  if (!Array.isArray(packages) || packages.length === 0) {
    return $t("common.appstore.noPackageRecords");
  }
  return packages
    .map((item) => packageTargetText(item))
    .join($t("common.appstore.listSeparator"));
};

const statusText = (status) =>
  status === 1 ? $t("common.appstore.enabled") : $t("common.appstore.disabled");
const yesNoText = (value) =>
  value === 1 ? $t("common.appstore.yes") : $t("common.appstore.no");
const isInstallable = (app) => app?.Installable !== false;
const installReasonText = (app) =>
  app?.InstallReason || $t("common.appstore.installReasonFallback");
const isInstalling = computed(() => !!installState.value);
const installProgressText = computed(() => {
  const progress = Number(installState.value?.progress) || 0;
  return `${progress}%`;
});

const detailScreenshots = computed(() => selectedApp.value?.Screenshots || []);

const detailCommonFields = computed(() => {
  if (!selectedApp.value) {
    return [];
  }
  const fields = [
    {
      label: $t("common.appstore.fields.code"),
      value: selectedApp.value.Code || "-",
    },
    {
      label: $t("common.appstore.fields.type"),
      value: pluginTypeText(selectedApp.value.PluginType),
    },
    {
      label: $t("common.appstore.fields.storeVersion"),
      value: selectedApp.value.StoreVersion || "-",
    },
    {
      label: $t("common.appstore.fields.minHostVersion"),
      value: selectedApp.value.LimitVersion || "-",
    },
  ];
  if (selectedApp.value.Installed) {
    fields.splice(3, 0, {
      label: $t("common.appstore.fields.installedVersion"),
      value: selectedApp.value.Version || "-",
    });
    if (selectedApp.value.Remark) {
      fields.push({
        label: $t("common.appstore.fields.remark"),
        value: selectedApp.value.Remark,
      });
    }
  }
  return fields;
});

const detailTypeFields = computed(() => {
  if (!selectedApp.value) {
    return [];
  }
  if (!selectedApp.value.Installed) {
    return [];
  }
  if (selectedApp.value.PluginType === 1) {
    return [
      {
        label: $t("common.appstore.fields.interceptors"),
        value: selectedApp.value.Interceptors || "-",
      },
      {
        label: $t("common.appstore.fields.debugPort"),
        value: selectedApp.value.DebugPort || "-",
      },
      {
        label: $t("common.appstore.fields.webhook"),
        value: selectedApp.value.WebHook || "-",
      },
      {
        label: $t("common.appstore.fields.closeDelay"),
        value: selectedApp.value.CloseDelay
          ? $t("common.appstore.minutes", {
              count: selectedApp.value.CloseDelay,
            })
          : $t("common.appstore.alwaysRunning"),
      },
      {
        label: $t("common.appstore.fields.needLogin"),
        value: yesNoText(selectedApp.value.NeedLogin),
      },
      {
        label: $t("common.appstore.fields.status"),
        value: statusText(selectedApp.value.Status),
      },
    ];
  }
  if (selectedApp.value.PluginType === 3) {
    const fields = [
      {
        label: $t("common.appstore.fields.firstMachine"),
        value: selectedApp.value.FirstMachine || "-",
      },
      {
        label: $t("common.appstore.fields.closeDelay"),
        value: selectedApp.value.CloseDelay
          ? $t("common.appstore.minutes", {
              count: selectedApp.value.CloseDelay,
            })
          : $t("common.appstore.alwaysRunning"),
      },
      {
        label: $t("common.appstore.fields.status"),
        value: statusText(selectedApp.value.Status),
      },
      {
        label: $t("common.appstore.fields.needLogin"),
        value: yesNoText(selectedApp.value.NeedLogin),
      },
    ];
    if (selectedApp.value.RuntimeError) {
      fields.push({
        label: "运行错误",
        value: selectedApp.value.RuntimeError,
      });
    }
    return fields;
  }
  if (selectedApp.value.PluginType === 4) {
    return [
      {
        label: "访问地址",
        value: selectedApp.value.AccessUrl || "-",
      },
      {
        label: $t("common.appstore.fields.status"),
        value: statusText(selectedApp.value.Status),
      },
    ];
  }
  return [];
});

const detailPackages = computed(() => selectedApp.value?.Packages || []);

const hasUpgrade = (app) => {
  if (!app?.Installed || !app.StoreVersion || !app.Version) {
    return false;
  }
  return compareVersions(app.StoreVersion, app.Version) > 0;
};

const loadDetail = async (code) => {
  detailLoading.value = true;
  const data = await postData("appstore", "detail", { code });
  if(data && data.app && data.app.iconUrl && data.app.iconUrl.startsWith("/")){
    data.app.iconUrl = await getApiHost() + data.app.iconUrl;
  }
  selectedApp.value = data ? normalizeApp(data.app || data) : null;
  appVersions.value = Array.isArray(data?.versions)
    ? data.versions.map(normalizeVersion)
    : [];
  detailLoading.value = false;
};

const refreshAfterMutation = async (code) => {
  await loadDetail(code);
  await hooks.refresh();
  await user.loadUser();
  emitter.emit("appstore-refresh", { code });
};

const installOrUpgradeApp = async (type, code) => {
  if (isInstalling.value) {
    return;
  }
  installState.value = { progress: 1 };
  let url = new URL(await getApiHost());
  const hostUrl = `${url.protocol}//${url.hostname}`;
  let success = false;
  let failed = false;
  try {
    const result = await postStreamData(
      "appstore",
      "install",
      {
        type,
        code,
        hostUrl,
      },
      (data) => {
        if (data?.code === -1) {
          failed = true;
          $msg.message.error(data.msg);
          return;
        }
        if (data?.progress !== undefined) {
          const progress = Math.max(1, Math.min(100, Number(data.progress) || 1));
          installState.value = { progress };
        }
        if (data?.done === true && data?.code === 0) {
          success = true;
        }
      },
      "okerr",
    );
    if (result && success && !failed) {
      await refreshAfterMutation(code);
    }
  } catch (err) {
    $msg.message.error(err?.message || String(err));
  } finally {
    installState.value = null;
  }
};

const installApp = async (code) => {
  await installOrUpgradeApp("1", code);
};

const uninstallApp = async (code) => {
  const result = await postData("appstore", "uninstall", { code }, "okerr");
  if (result) {
    await refreshAfterMutation(code);
  }
};

const upgradeApp = async (code) => {
  await installOrUpgradeApp("2", code);
};

const openPlugin = async (app) => {
  const item = normalizeApp(app);
  let path = `/page/${item.Code}/`;
  if (item.PluginType === 2) {
    path = `/api/webplugin/${item.Code}/`;
  } else if (item.PluginType === 3) {
    path = `/api/thirdplugin/${item.Code}/`;
  } else if (item.PluginType === 4) {
    path = item.AccessUrl;
  }
  const host = await getApiHost();
  const size = getWinSize();
  $wins.addWindow(
    {
      icon:app.IconUrl,
      width: size.width,
      height: size.height,
      title: item.Name || item.Code,
      iframeUrl: item.PluginType === 4 ? path : `${host}${path}`,
    }
  );
};

const prevScreenshot = () => {
  if (screenshotIndex.value > 0) {
    screenshotIndex.value -= 1;
  }
};

const nextScreenshot = () => {
  if (screenshotIndex.value < detailScreenshots.value.length - 1) {
    screenshotIndex.value += 1;
  }
};

watch(selectedApp, () => {
  screenshotIndex.value = 0;
});

watch(
  () => String(props.data?.code || props.data?.Code || "").trim(),
  async (code) => {
    if (!code) {
      selectedApp.value = null;
      appVersions.value = [];
      return;
    }
    await loadDetail(code);
  },
  { immediate: true },
);
</script>

<template>
  <n-spin :show="detailLoading">
    <div v-if="selectedApp" class="min-h-full p-5 user-color-ftext">
      <div class="flex gap-5 mb-6">
        <div>
          <img
            v-if="selectedApp.IconUrl"
            :src="selectedApp.IconUrl"
            class="w-24 h-24 object-cover user-rounded-5"
          />
          <div v-else class="w-24 h-24 flex items-center justify-center user-rounded-5 user-color-surface-muted text-9 font-700">?</div>
        </div>

        <div class="flex-1 min-w-0">
          <h2 class="m-0 text-6 leading-1.25">
            {{ selectedApp.Name || selectedApp.Code }}
          </h2>
          <div class="mt-2.5 text-3.25 opacity-60">
            {{ selectedApp.Author || pluginTypeText(selectedApp.PluginType) }}
          </div>
          <div class="text-3 opacity-50">
            {{ $t("common.appstore.storeVersion") }}
            {{ selectedApp.StoreVersion || "-" }}
            <span v-if="selectedApp.Installed">
              / {{ $t("common.appstore.installedVersion") }}
              {{ selectedApp.Version || "-" }}</span
            >
          </div>
          <n-space class="mt-3.5" :size="8">
            <n-button
              v-if="canInstall && !selectedApp.Installed"
              type="primary"
              round
              :loading="isInstalling"
              :disabled="!isInstallable(selectedApp) || isInstalling"
              @click="installApp(selectedApp.Code)"
            >
              {{
                isInstalling
                  ? installProgressText
                  : isInstallable(selectedApp)
                  ? $t("common.appstore.install")
                  : $t("common.appstore.incompatible")
              }}
            </n-button>
            <n-button
              v-if="canUpgrade && hasUpgrade(selectedApp)"
              type="info"
              round
              :loading="isInstalling"
              :disabled="isInstalling"
              @click="upgradeApp(selectedApp.Code)"
            >
              {{ isInstalling ? installProgressText : $t("common.appstore.upgrade") }}
            </n-button>
            <n-button
              v-if="selectedApp.Installed"
              round
              @click="openPlugin(selectedApp)"
            >
              {{ $t("common.appstore.open") }}
            </n-button>
            <n-button
              v-if="canUninstall && selectedApp.Installed"
              type="error"
              round
              @click="uninstallApp(selectedApp.Code)"
            >
              {{ $t("common.appstore.uninstall") }}
            </n-button>
          </n-space>
          <div
            v-if="!selectedApp.Installed && !isInstallable(selectedApp)"
            class="mt-2.5 text-3 leading-1.6 text-[var(--n-error-color)]"
          >
            {{ installReasonText(selectedApp) }}
          </div>
        </div>
      </div>

      <div v-if="detailScreenshots.length > 0" class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">
          {{ $t("common.appstore.screenshotPreview") }}
        </h3>
        <div class="flex items-center gap-2.5">
          <button
            v-if="detailScreenshots.length > 1"
            class="w-8.5 h-8.5 border user-color-border user-rounded-full user-color-fbg user-color-ftext cursor-pointer disabled:opacity-35 disabled:cursor-default"
            :disabled="screenshotIndex === 0"
            @click="prevScreenshot"
          >
            &lt;
          </button>
          <div class="flex-1 min-h-55 flex items-center justify-center border user-color-border user-rounded-2.5 user-color-surface-muted">
            <n-image
              :src="detailScreenshots[screenshotIndex]"
              height="220"
              object-fit="contain"
            />
          </div>
          <button
            v-if="detailScreenshots.length > 1"
            class="w-8.5 h-8.5 border user-color-border user-rounded-full user-color-fbg user-color-ftext cursor-pointer disabled:opacity-35 disabled:cursor-default"
            :disabled="screenshotIndex === detailScreenshots.length - 1"
            @click="nextScreenshot"
          >
            &gt;
          </button>
        </div>
      </div>

      <div class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">{{ $t("common.appstore.summary") }}</h3>
        <p class="m-0 whitespace-pre-wrap leading-1.7 text-3.25 opacity-82">
          {{ selectedApp.Description || $t("common.appstore.noDescription") }}
        </p>
      </div>

      <div class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">{{ $t("common.appstore.basicInfo") }}</h3>
        <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
          <div
            v-for="field in detailCommonFields"
            :key="field.label"
            class="px-3.5 py-3 border user-color-border user-rounded-2.5 user-color-surface-muted"
          >
            <div class="text-3 opacity-55">{{ field.label }}</div>
            <div class="mt-1.5 text-3.25 leading-1.5">{{ field.value }}</div>
          </div>
        </div>
      </div>

      <div v-if="detailTypeFields.length > 0" class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">
          {{ $t("common.appstore.typeCapabilities") }}
        </h3>
        <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
          <div
            v-for="field in detailTypeFields"
            :key="field.label"
            class="px-3.5 py-3 border user-color-border user-rounded-2.5 user-color-surface-muted"
          >
            <div class="text-3 opacity-55">{{ field.label }}</div>
            <div class="mt-1.5 text-3.25 leading-1.5 break-all">{{ field.value }}</div>
          </div>
        </div>
      </div>

      <div v-if="detailPackages.length > 0" class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">
          {{ $t("common.appstore.availablePackages") }}
        </h3>
        <div class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
          <div
            v-for="pkg in detailPackages"
            :key="`${pkg.OS}-${pkg.Arch}-${pkg.PackageName}`"
            class="px-3.5 py-3 border user-color-border user-rounded-2.5 user-color-surface-muted"
          >
            <div class="text-3 opacity-55">{{ packageTargetText(pkg) }}</div>
            <div class="mt-1.5 text-3.25 leading-1.5 break-all">
              {{ pkg.PackageName || "-" }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="appVersions.length > 0" class="mb-6">
        <h3 class="m-0 mb-3 text-3.75 font-600">{{ $t("common.appstore.versionHistory") }}</h3>
        <div
          v-for="version in appVersions"
          :key="version.Version"
          class="py-3 border-b user-color-border last:border-b-0"
        >
          <div class="flex justify-between gap-3">
            <span class="font-600">{{ version.Version || "-" }}</span>
            <span class="text-3 opacity-48">{{ version.CreatedAt || "-" }}</span>
          </div>
          <div class="mt-1.5 text-3 opacity-74 whitespace-pre-wrap">
            {{ version.Changelog || $t("common.appstore.noUpdateNotes") }}
          </div>
          <div class="mt-1.5 text-3 opacity-60">
            {{ versionPackageSummary(version.Packages) }}
          </div>
        </div>
      </div>
    </div>

    <div v-else class="min-h-60 flex items-center justify-center">
      <n-empty :description="$t('common.appstore.loadFailed')" />
    </div>
  </n-spin>
</template>
