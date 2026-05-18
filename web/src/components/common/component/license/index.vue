<script setup>
import { defineAsyncComponent, h, onMounted, ref } from "vue";
import { NAlert, NButton, NCard, NDataTable, NDescriptions, NDescriptionsItem, NEmpty, NSpace, NTag } from "naive-ui";
import { postBlob, postData } from "@/util/util";
import { checkAuth } from "@/directives/auth";

let OfficialLicenseEe = null;
const mod = window.__INFACE_MODS__?.["/src/components/common/inface/official_license_ee/index.vue"];
if (mod) {
  OfficialLicenseEe = defineAsyncComponent(mod);
}

const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});

const loading = ref(false);
const license = ref({ features: [] });

const authTypeText = (value) => ({ has: "永久授权", times: "数量授权", expired: "到期授权" }[value] || value || "-");
const statusText = (value) => ({ valid: "有效", expired: "已过期", disabled: "未启用" }[value] || value || "-");

function formatTime(ts) {
  const value = Number(ts || 0);
  if (value <= 0) return "-";
  return new Date(value * 1000).toLocaleString();
}

const columns = [
  { title: "项目编码", key: "code", minWidth: 180 },
  { title: "项目名称", key: "name", minWidth: 180, render: (row) => row.name || "-" },
  { title: "授权类型", key: "authType", width: 120, render: (row) => authTypeText(row.authType) },
  { title: "数量", key: "quantity", width: 110, render: (row) => row.quantity || "-" },
  { title: "到期时间", key: "expiredAt", width: 170, render: (row) => formatTime(row.expiredAt) },
  { title: "插件密文", key: "hasCipher", width: 110, render: (row) => (row.hasCipher ? "是" : "否") },
  {
    title: "状态",
    key: "status",
    width: 100,
    render: (row) => h(NTag, { size: "small", type: row.status === "valid" ? "success" : "warning" }, { default: () => statusText(row.status) }),
  },
];

async function loadData() {
  loading.value = true;
  const res = await postData("license", "page", {}, "");
  loading.value = false;
  license.value = res?.license || { features: [] };
}

async function exportLicense() {
  const data = await postBlob("license", "export", {}, "okerr");
  if (!data) {
    return;
  }
  const a = document.createElement("a");
  a.href = URL.createObjectURL(data.blob);
  a.download = data.name || "插件授权信息.v9os";
  a.click();
  URL.revokeObjectURL(a.href);
}

async function importLicense(evt) {
  const file = evt.target.files?.[0];
  evt.target.value = "";
  if (!file) return;
  const formData = new FormData();
  formData.append("file", file);
  const res = await postData("license", "import", formData, "okerr");
  if (res) await loadData();
}

async function abandonLicense() {
  const ok = await $msg.util.confirm("放弃本地插件授权后，需要重新导入授权文件。", "放弃授权");
  if (!ok) return;
  const res = await postData("license", "abandon", {}, "okerr");
  if (res) await loadData();
}

onMounted(() => {
  if (!OfficialLicenseEe) {
    loadData();
  }
});
</script>

<template>
  <OfficialLicenseEe v-if="OfficialLicenseEe" :win-id="props.winId" />
  <div v-else-if="!checkAuth('/api/license/page')" class="flex items-center justify-center h-full">
    <n-empty description="无权限访问" />
  </div>
  <div v-else class="h-full overflow-auto p-4">
    <div>
      <div class="flex items-center justify-between gap-3 mb-4">
        <div>
          <div class="text-18px font-700">插件商业授权</div>
          <div class="text-12px opacity-70 mt-1">源码版进行插件授权专用。</div>
        </div>
        <n-space>
          <n-button size="small" :loading="loading" @click="loadData">刷新</n-button>
          <n-button size="small" @click="exportLicense">导出</n-button>
          <n-button size="small" type="primary" tag="label">
            导入
            <input type="file" accept=".json,.v9os,application/json" class="hidden" @change="importLicense" />
          </n-button>
          <n-button size="small" type="error" secondary @click="abandonLicense">放弃授权</n-button>
        </n-space>
      </div>

      <n-alert v-if="!license.authorized" type="warning" class="mb-4">
        {{ license.unavailableText || "当前节点未安装有效插件商业授权" }}
      </n-alert>

      <n-card size="small" class="mb-4">
        <n-descriptions :column="2" label-placement="left" size="small">
          <n-descriptions-item label="授权 ID">{{ license.authId || "-" }}</n-descriptions-item>
          <n-descriptions-item label="授权项目">{{ license.featureCount || 0 }}</n-descriptions-item>
          <n-descriptions-item label="开始时间">{{ formatTime(license.startAt) }}</n-descriptions-item>
          <n-descriptions-item label="结束时间">{{ formatTime(license.endAt) }}</n-descriptions-item>
        </n-descriptions>
      </n-card>

      <n-card size="small">
        <template #header>授权项目</template>
        <n-data-table :columns="columns" :data="license.features || []" :loading="loading" :pagination="{ pageSize: 10 }" />
      </n-card>
    </div>
  </div>
</template>
