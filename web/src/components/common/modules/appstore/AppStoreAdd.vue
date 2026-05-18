<script setup>
import { computed, ref } from "vue";
import {
  NButton,
  NInput,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
} from "naive-ui";
import { windowsStore } from "@/stores/windows";
import { useStore } from "@/stores/user.js";
import { webhookStore } from "@/stores/webhook.js";
import emitter from "@/util/event.js";
import { getApiHost, postData } from "@/util/util";

const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});

const ws = windowsStore();
const user = useStore();
const hooks = webhookStore();

const pluginType = ref(4);
const osName = ref("");
const arch = ref("");
const selectedFile = ref(null);
const saving = ref(false);
const fileInput = ref(null);
const form = ref({
  name: "",
  accessUrl: "",
  iconUrl: "",
  description: "",
  remark: ""
});

const typeOptions = computed(() => [
  { label: $t("common.appstore.pluginTypes.main"), value: 1 },
  { label: $t("common.appstore.pluginTypes.frontend"), value: 2 },
  { label: $t("common.appstore.pluginTypes.thirdParty"), value: 3 },
  { label: $t("common.appstore.pluginTypes.remote"), value: 4 },
]);

const isRemote = computed(() => pluginType.value === 4);
const needTarget = computed(() => pluginType.value === 1 || pluginType.value === 3);
const fileName = computed(() => selectedFile.value?.name || $t("common.appstore.noFileSelected"));

const osOptions = [
  { label: "windows", value: "windows" },
  { label: "linux", value: "linux" },
  { label: "darwin", value: "darwin" },
  { label: "android", value: "android" },
];

const archOptions = [
  { label: "amd64", value: "amd64" },
  { label: "arm64", value: "arm64" },
];

const closeWindow = () => {
  if (props.winId) ws.closeWindow(props.winId);
};

const chooseFile = () => {
  fileInput.value?.click();
};

const handleFileChange = (event) => {
  selectedFile.value = event.target.files?.[0] || null;
  event.target.value = "";
};

const submit = async () => {
  if (isRemote.value) {
    if (!form.value.accessUrl.trim()) {
      $msg.message.error($t("common.appstore.accessUrlRequired"));
      return;
    }
  } else if (!selectedFile.value) {
    $msg.message.error($t("common.appstore.packageRequired"));
    return;
  } else if (needTarget.value && (!osName.value || !arch.value)) {
    $msg.message.error($t("common.appstore.osArchRequired"));
    return;
  }
  saving.value = true;
  try {
    const data = new FormData();
    data.append("pluginType", String(pluginType.value));
    data.append("os", osName.value);
    data.append("arch", arch.value);
    if (selectedFile.value) {
      data.append("file", selectedFile.value);
    }
    const url = new URL(await getApiHost());
    data.append("hostUrl", `${url.protocol}//${url.hostname}`);
    data.append("name", form.value.name.trim());
    data.append("accessUrl", form.value.accessUrl.trim());
    data.append("iconUrl", form.value.iconUrl.trim());
    data.append("description", form.value.description.trim());
    data.append("remark", form.value.remark.trim());
    const result = await postData("appstore", "add", data, "okerr");
    if (result) {
      emitter.emit("appstore-refresh", {});
      await hooks.refresh();
      await user.loadUser();
      closeWindow();
    }
  } finally {
    saving.value = false;
  }
};
</script>

<template>
  <div class="flex h-full min-h-0 flex-col gap-4 p-5 user-color-ftext">
    <div class="grid gap-1.5">
      <div class="text-[12px] opacity-70">{{ $t("common.appstore.appType") }}</div>
      <n-select v-model:value="pluginType" :options="typeOptions" />
    </div>

    <template v-if="isRemote">
      <div class="grid gap-1.5">
        <div class="text-[12px] opacity-70">{{ $t("common.appstore.fields.name") }}</div>
        <n-input v-model:value="form.name" :placeholder="$t('common.appstore.fields.name')" />
      </div>
      <div class="grid gap-1.5">
        <div class="text-[12px] opacity-70">{{ $t("common.appstore.fields.accessUrl") }}</div>
        <n-input v-model:value="form.accessUrl" placeholder="https://example.com" />
      </div>
      <div class="grid gap-1.5">
        <div class="text-[12px] opacity-70">{{ $t("common.appstore.fields.iconUrl") }}</div>
        <n-input v-model:value="form.iconUrl" placeholder="https://example.com/icon.png" />
      </div>
      <div class="grid gap-1.5">
        <div class="text-[12px] opacity-70">{{ $t("common.appstore.summary") }}</div>
        <n-input v-model:value="form.description" type="textarea" :autosize="{ minRows: 3, maxRows: 5 }" />
      </div>
    </template>

    <template v-else>
      <div v-if="needTarget" class="grid grid-cols-2 gap-3">
        <div class="grid gap-1.5">
          <div class="text-[12px] opacity-70">{{ $t("common.appstore.os") }}</div>
          <n-select v-model:value="osName" :options="osOptions" />
        </div>
        <div class="grid gap-1.5">
          <div class="text-[12px] opacity-70">{{ $t("common.appstore.arch") }}</div>
          <n-select v-model:value="arch" :options="archOptions" />
        </div>
      </div>

      <div class="grid gap-2">
        <div class="text-[12px] opacity-70">{{ $t("common.appstore.localPackage") }}</div>
        <div class="flex items-center gap-2">
          <n-button @click="chooseFile">{{ $t("common.appstore.selectPackage") }}</n-button>
          <n-tag class="min-w-0 max-w-full">
            <span class="inline-block max-w-[360px] truncate align-bottom">{{ fileName }}</span>
          </n-tag>
        </div>
        <input ref="fileInput" type="file" class="hidden" accept=".zip" @change="handleFileChange" />
      </div>
    </template>

    <div class="mt-auto flex justify-end border-t user-color-border pt-3">
      <n-space>
        <n-button @click="closeWindow">{{ $t("common.all.cancel") }}</n-button>
        <n-button type="primary" :loading="saving" @click="submit">{{ $t("common.all.submit") }}</n-button>
      </n-space>
    </div>
  </div>
</template>
