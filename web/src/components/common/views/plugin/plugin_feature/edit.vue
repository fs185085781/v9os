<script setup>
import { NForm, NFormItem, NInput, NButton,NSelect,NDatePicker,NGrid,NGridItem } from 'naive-ui'
import { windowsStore } from "@/stores/windows"
import { computed } from "vue";
import { postData } from "@/util/util";
const props = defineProps({
    data: {
    },
    winId: {
        type: String,
        default: ""
    }
})
const row = props.data
const winId = props.winId
const ws = windowsStore()
import emitter from '@/util/event.js';
const save = async () => {
    const res = await postData("plugin_feature","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("plugin_feature-edit-over", {})
    }
}
const enabledOptions = computed(() => [
    {
        label: $t("model.plugin_feature.enabled_select_0"),
        value: "0"
    },{
        label: $t("model.plugin_feature.enabled_select_1"),
        value: "1"
    },
])
</script>
<template>
    <n-form
    :model="row"
    label-placement="left"
    label-width="auto"
    require-mark-placement="right-hanging"
    size="small"
    class="w-full p-5"
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.plugin_feature.plugin_code')" path="PluginCode">
      <n-input v-model:value="row.PluginCode" :placeholder="$t('model.plugin_feature.plugin_code')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_feature.enabled')" path="Enabled">
      <n-select v-model:value="row.Enabled" :options="enabledOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_feature.content')" path="Content">
      <n-input v-model:value="row.Content" :placeholder="$t('model.plugin_feature.content')" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>