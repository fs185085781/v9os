<script setup>
import { computed, reactive } from "vue";
import { NButton, NForm, NFormItem, NInput, NSelect } from "naive-ui";
import { windowsStore } from "@/stores/windows";

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

const ws = windowsStore();
const form = reactive({
  ID: props.data.ID || props.data.id || 0,
  Title: props.data.Title || "",
  Icon: props.data.Icon || $user.defaultIcon("appstore"),
  AppType: props.data.AppType || "iframe",
  Code: props.data.Code || "",
  Url: props.data.Url || "",
  Sort: Number(props.data.Sort || props.data.sort || 0),
});

const typeOptions = computed(() => [
  { label: $t("model.desktop_app.app_type_select_system"), value: "system" },
  { label: $t("model.desktop_app.app_type_select_plugin"), value: "plugin" },
  { label: $t("model.desktop_app.app_type_select_iframe"), value: "iframe" },
]);
const showIcon = computed(() => form.AppType !== "system");
const showCode = computed(() => form.AppType !== "iframe");
const showUrl = computed(() => form.AppType !== "system");

async function save() {
  if (!form.Title.trim()) {
    $msg.message.error($t("common.desktop_app.title_required"));
    return;
  }
  if (showCode.value && !form.Code.trim()) {
    $msg.message.error($t("common.desktop_app.code_required"));
    return;
  }
  if (form.AppType === "iframe" && !form.Url.trim()) {
    $msg.message.error($t("common.desktop_app.url_required"));
    return;
  }
  const data = { ...form };
  if (data.AppType === "system") {
    data.Icon = "";
    data.Url = "";
  }
  if (data.AppType === "iframe") {
    data.Code = "";
  }
  const res = await $user.saveDesktopApp(data);
  if (res && props.winId) {
    ws.closeWindow(props.winId);
  }
}
</script>

<template>
  <div class="h-full p-4">
    <n-form label-placement="left" label-width="86">
      <n-form-item :label="$t('model.desktop_app.title')">
        <n-input v-model:value="form.Title" :placeholder="$t('model.desktop_app.title')" />
      </n-form-item>
      <n-form-item v-if="showIcon" :label="$t('model.desktop_app.icon')">
        <n-input v-model:value="form.Icon" :placeholder="$t('model.desktop_app.icon')" />
      </n-form-item>
      <n-form-item :label="$t('model.desktop_app.app_type')">
        <n-select v-model:value="form.AppType" :options="typeOptions" />
      </n-form-item>
      <n-form-item v-if="showCode" :label="$t('model.desktop_app.code')">
        <n-input v-model:value="form.Code" :placeholder="$t('model.desktop_app.code')" />
      </n-form-item>
      <n-form-item v-if="showUrl" :label="$t('model.desktop_app.url')">
        <n-input v-model:value="form.Url" :placeholder="$t('model.desktop_app.url')" />
      </n-form-item>
      <div class="flex justify-end gap-2">
        <n-button @click="props.winId && ws.closeWindow(props.winId)">{{ $t("common.all.cancel") }}</n-button>
        <n-button type="primary" @click="save">{{ $t("common.all.save") }}</n-button>
      </div>
    </n-form>
  </div>
</template>
