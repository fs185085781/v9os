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
    const res = await postData("plugin","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("plugin-edit-over", {})
    }
}
const statusOptions = computed(() => [
    {
        label: $t("model.plugin.status_select_0"),
        value: "0"
    },{
        label: $t("model.plugin.status_select_1"),
        value: "1"
    },
])
const plugin_typeOptions = computed(() => [
    {
        label: $t("model.plugin.plugin_type_select_1"),
        value: "1"
    },{
        label: $t("model.plugin.plugin_type_select_2"),
        value: "2"
    },{
        label: $t("model.plugin.plugin_type_select_3"),
        value: "3"
    },{
        label: $t("model.plugin.plugin_type_select_4"),
        value: "4"
    },
])
const need_loginOptions = computed(() => [
    {
        label: $t("model.plugin.need_login_select_0"),
        value: "0"
    },{
        label: $t("model.plugin.need_login_select_1"),
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.plugin.first_machine')" path="FirstMachine">
      <n-input v-model:value="row.FirstMachine" :placeholder="$t('model.plugin.first_machine')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.runtime_error')" path="RuntimeError">
      <n-input v-model:value="row.RuntimeError" :placeholder="$t('model.plugin.runtime_error')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.name')" path="Name">
      <n-input v-model:value="row.Name" :placeholder="$t('model.plugin.name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.description')" path="Description">
      <n-input type="textarea" v-model:value="row.Description" :placeholder="$t('model.plugin.description')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.close_delay')" path="CloseDelay">
      <n-input v-model:value="row.CloseDelay" :placeholder="$t('model.plugin.close_delay')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.code')" path="Code">
      <n-input v-model:value="row.Code" :placeholder="$t('model.plugin.code')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.status')" path="Status">
      <n-select v-model:value="row.Status" :options="statusOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.remark')" path="Remark">
      <n-input v-model:value="row.Remark" :placeholder="$t('model.plugin.remark')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.version')" path="Version">
      <n-input v-model:value="row.Version" :placeholder="$t('model.plugin.version')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.plugin_type')" path="PluginType">
      <n-select v-model:value="row.PluginType" :options="plugin_typeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.interceptors')" path="Interceptors">
      <n-input v-model:value="row.Interceptors" :placeholder="$t('model.plugin.interceptors')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.web_hook')" path="WebHook">
      <n-input v-model:value="row.WebHook" :placeholder="$t('model.plugin.web_hook')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.limit_version')" path="LimitVersion">
      <n-input v-model:value="row.LimitVersion" :placeholder="$t('model.plugin.limit_version')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.need_login')" path="NeedLogin">
      <n-select v-model:value="row.NeedLogin" :options="need_loginOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.icon_url')" path="IconUrl">
      <n-input v-model:value="row.IconUrl" :placeholder="$t('model.plugin.icon_url')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.access_url')" path="AccessUrl">
      <n-input v-model:value="row.AccessUrl" :placeholder="$t('model.plugin.access_url')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin.debug_port')" path="DebugPort">
      <n-input v-model:value="row.DebugPort" :placeholder="$t('model.plugin.debug_port')" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>