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
    const res = await postData("plugin_web_data","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("plugin_web_data-edit-over", {})
    }
}
</script>
<template>
    <n-form
    :model="row"
    label-placement="left"
    label-width="auto"
    require-mark-placement="right-hanging"
    size="small"
    class="w-full p-5"
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.plugin_web_data.code')" path="Code">
      <n-input v-model:value="row.Code" :placeholder="$t('model.plugin_web_data.code')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_web_data.user_id')" path="UserID">
      <n-input v-model:value="row.UserID" :placeholder="$t('model.plugin_web_data.user_id')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_web_data.data_key')" path="DataKey">
      <n-input v-model:value="row.DataKey" :placeholder="$t('model.plugin_web_data.data_key')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_web_data.data_value')" path="DataValue">
      <n-input type="textarea" v-model:value="row.DataValue" :placeholder="$t('model.plugin_web_data.data_value')" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>