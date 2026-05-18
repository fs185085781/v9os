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
    const res = await postData("plugin_column","save",row)
    if(res){
        ws.closeWindow(winId)
        emitter.emit("plugin_column-edit-over", {})
    }
}
const field_typeOptions = computed(() => [
    {
        label: $t("model.plugin_column.field_type_select_float64"),
        value: "float64"
    },{
        label: $t("model.plugin_column.field_type_select_int"),
        value: "int"
    },{
        label: $t("model.plugin_column.field_type_select_int64"),
        value: "int64"
    },{
        label: $t("model.plugin_column.field_type_select_string"),
        value: "string"
    },
])
const is_indexOptions = computed(() => [
    {
        label: $t("model.plugin_column.is_index_select_1"),
        value: "1"
    },{
        label: $t("model.plugin_column.is_index_select_2"),
        value: "2"
    },
])
const is_textOptions = computed(() => [
    {
        label: $t("model.plugin_column.is_text_select_1"),
        value: "1"
    },{
        label: $t("model.plugin_column.is_text_select_2"),
        value: "2"
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.plugin_column.plugin_name')" path="PluginName">
      <n-input v-model:value="row.PluginName" :placeholder="$t('model.plugin_column.plugin_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.plugin_table')" path="PluginTable">
      <n-input v-model:value="row.PluginTable" :placeholder="$t('model.plugin_column.plugin_table')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.main_field_name')" path="MainFieldName">
      <n-input v-model:value="row.MainFieldName" :placeholder="$t('model.plugin_column.main_field_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.main_column_name')" path="MainColumnName">
      <n-input v-model:value="row.MainColumnName" :placeholder="$t('model.plugin_column.main_column_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.field_name')" path="FieldName">
      <n-input v-model:value="row.FieldName" :placeholder="$t('model.plugin_column.field_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.column_name')" path="ColumnName">
      <n-input v-model:value="row.ColumnName" :placeholder="$t('model.plugin_column.column_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.field_type')" path="FieldType">
      <n-select v-model:value="row.FieldType" :options="field_typeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.is_index')" path="IsIndex">
      <n-select v-model:value="row.IsIndex" :options="is_indexOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_column.is_text')" path="IsText">
      <n-select v-model:value="row.IsText" :options="is_textOptions" />
    </n-form-item></n-grid-item>
    
  </n-grid>
  <div class="flex justify-end">
      <n-button type="primary" @click="save()" size="small">
        {{ $t("common.all.save") }}
      </n-button>
  </div>
  </n-form>
</template>