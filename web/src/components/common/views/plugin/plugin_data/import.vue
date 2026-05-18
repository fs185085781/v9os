<script setup>
import { NGrid, NGi, NFlex, NButton, NAlert } from "naive-ui";
import { reactive, ref, computed } from "vue";
import { postBlob, postData } from "@/util/util";
const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});
const table = props.data.Table;
const data = reactive({
  title: $tc("common.all.waitingImport"),
  type: "info",
  content: $tc("common.all.waitingImportContent"),
});
const downloadTemplate = async () => {
  const data = await postBlob("plugin_data", "import?table=" + table, {
    isTemplate: true,
  });
  if (data == null) {
    return;
  }
  const a = document.createElement("a");
  a.href = URL.createObjectURL(data.blob);
  a.download = "import.xlsx";
  a.click();
};
const fileInput = ref();
const importTemplate = async () => {
  fileInput.value.click();
};
import emitter from "@/util/event.js";
const handleFileUpload = async (event) => {
  const selectedFile = event.target.files[0];
  if (!selectedFile) {
    return;
  }
  event.target.value = "";
  const formData = new FormData();
  formData.append("file", selectedFile);
  let msg = "";
  let flag = await postData(
    "plugin_data",
    "import?table=" + table,
    formData,
    "",
    function (res) {
      msg = res;
    },
  );
  if (flag) {
    data.title = $tc("common.all.importSuccess");
    data.type = "success";
    data.content = computed(() => {
      return msg;
    });
    emitter.emit("plugin_data-edit-over", {});
  } else {
    data.title = $tc("common.all.importFailed");
    data.type = "error";
    data.content = computed(() => {
      return msg;
    });
  }
};
</script>
<template>
  <n-grid :cols="1" class="p-2">
    <n-gi>
      <n-flex inline>
        <n-button type="primary" size="small" @click="downloadTemplate">
          {{ $t("common.all.downloadTemplate") }}
        </n-button>
        <n-button type="primary" size="small" @click="importTemplate">
          {{ $t("common.all.importTemplate") }}
        </n-button>
      </n-flex>
    </n-gi>
    <n-gi class="mt-2">
      <n-alert :title="data.title" :type="data.type">
        <div>{{ data.content }}</div>
      </n-alert>
    </n-gi>
  </n-grid>
  <input
    ref="fileInput"
    type="file"
    class="hidden"
    accept=".xlsx"
    @change="handleFileUpload"
  />
</template>
