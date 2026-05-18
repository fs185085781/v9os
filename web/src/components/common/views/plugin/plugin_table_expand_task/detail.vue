<script setup>
import { NForm, NFormItem, NInput,NSelect,NDatePicker,NGrid,NGridItem } from 'naive-ui'
import { computed } from "vue";
const props = defineProps({
    data: {
    },
    winId: {
        type: String,
        default: ""
    }
})
const row = props.data
const statusOptions = computed(() => [
    {
        label: $t("model.plugin_table_expand_task.status_select_1"),
        value: "1"
    },{
        label: $t("model.plugin_table_expand_task.status_select_2"),
        value: "2"
    },{
        label: $t("model.plugin_table_expand_task.status_select_3"),
        value: "3"
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
    :disabled="true"
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.plugin_table_id')" path="PluginTableID">
      <n-input v-model:value="row.PluginTableID" :placeholder="$t('model.plugin_table_expand_task.plugin_table_id')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.plugin_name')" path="PluginName">
      <n-input v-model:value="row.PluginName" :placeholder="$t('model.plugin_table_expand_task.plugin_name')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.plugin_table')" path="PluginTable">
      <n-input v-model:value="row.PluginTable" :placeholder="$t('model.plugin_table_expand_task.plugin_table')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.old_real_table')" path="OldRealTable">
      <n-input v-model:value="row.OldRealTable" :placeholder="$t('model.plugin_table_expand_task.old_real_table')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.new_real_table')" path="NewRealTable">
      <n-input v-model:value="row.NewRealTable" :placeholder="$t('model.plugin_table_expand_task.new_real_table')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.data_length')" path="DataLength">
      <n-input v-model:value="row.DataLength" :placeholder="$t('model.plugin_table_expand_task.data_length')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.migrated_rows')" path="MigratedRows">
      <n-input v-model:value="row.MigratedRows" :placeholder="$t('model.plugin_table_expand_task.migrated_rows')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.retry_count')" path="RetryCount">
      <n-input v-model:value="row.RetryCount" :placeholder="$t('model.plugin_table_expand_task.retry_count')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.status')" path="Status">
      <n-select v-model:value="row.Status" :options="statusOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.plugin_table_expand_task.error_msg')" path="ErrorMsg">
      <n-input v-model:value="row.ErrorMsg" :placeholder="$t('model.plugin_table_expand_task.error_msg')" />
    </n-form-item></n-grid-item>
    
    <n-grid-item><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
      <n-date-picker class="w-full" v-model:value="row.CreatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
      <n-date-picker class="w-full" v-model:value="row.UpdatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    </n-grid></n-form>
</template>