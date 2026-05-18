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
    const res = await postData("log","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("log-edit-over", {})
    }
}
const levelOptions = computed(() => [
    {
        label: $t("model.log.level_select_debug"),
        value: "debug"
    },{
        label: $t("model.log.level_select_error"),
        value: "error"
    },{
        label: $t("model.log.level_select_info"),
        value: "info"
    },{
        label: $t("model.log.level_select_warn"),
        value: "warn"
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.log.level')" path="Level">
      <n-select v-model:value="row.Level" :options="levelOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.log.msg')" path="Msg">
      <n-input v-model:value="row.Msg" :placeholder="$t('model.log.msg')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.log.time')" path="Time">
      <n-date-picker class="w-full" v-model:value="row.Time" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.log.text')" path="Text">
      <n-input type="textarea" v-model:value="row.Text" :placeholder="$t('model.log.text')" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>