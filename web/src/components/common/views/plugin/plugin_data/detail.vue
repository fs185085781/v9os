<script setup>
import {
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NDatePicker,
  NGrid,
  NGridItem,
} from "naive-ui";
import { computed } from "vue";
const props = defineProps({
  data: {},
  winId: {
    type: String,
    default: "",
  },
});
const row = props.data;
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
    ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"
      ><n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.plugin_name')"
          path="PluginName"
        >
          <n-input
            v-model:value="row.PluginName"
            :placeholder="$t('model.plugin_data.plugin_name')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.plugin_table')"
          path="PluginTable"
        >
          <n-input
            v-model:value="row.PluginTable"
            :placeholder="$t('model.plugin_data.plugin_table')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.data_id')" path="DataId">
          <n-input
            v-model:value="row.DataId"
            :placeholder="$t('model.plugin_data.data_id')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.user_id')" path="UserId">
          <n-input
            v-model:value="row.UserId"
            :placeholder="$t('model.plugin_data.user_id')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.dept_id')" path="DeptId">
          <n-input
            v-model:value="row.DeptId"
            :placeholder="$t('model.plugin_data.dept_id')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field1')" path="Field1">
          <n-input
            v-model:value="row.Field1"
            :placeholder="$t('model.plugin_data.field1')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field2')" path="Field2">
          <n-input
            v-model:value="row.Field2"
            :placeholder="$t('model.plugin_data.field2')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field3')" path="Field3">
          <n-input
            v-model:value="row.Field3"
            :placeholder="$t('model.plugin_data.field3')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field4')" path="Field4">
          <n-input
            v-model:value="row.Field4"
            :placeholder="$t('model.plugin_data.field4')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field5')" path="Field5">
          <n-input
            v-model:value="row.Field5"
            :placeholder="$t('model.plugin_data.field5')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field6')" path="Field6">
          <n-input
            v-model:value="row.Field6"
            :placeholder="$t('model.plugin_data.field6')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field7')" path="Field7">
          <n-input
            v-model:value="row.Field7"
            :placeholder="$t('model.plugin_data.field7')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field8')" path="Field8">
          <n-input
            v-model:value="row.Field8"
            :placeholder="$t('model.plugin_data.field8')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field9')" path="Field9">
          <n-input
            v-model:value="row.Field9"
            :placeholder="$t('model.plugin_data.field9')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field10')" path="Field10">
          <n-input
            v-model:value="row.Field10"
            :placeholder="$t('model.plugin_data.field10')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field11')" path="Field11">
          <n-input
            v-model:value="row.Field11"
            :placeholder="$t('model.plugin_data.field11')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field12')" path="Field12">
          <n-input
            v-model:value="row.Field12"
            :placeholder="$t('model.plugin_data.field12')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field13')" path="Field13">
          <n-input
            v-model:value="row.Field13"
            :placeholder="$t('model.plugin_data.field13')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field14')" path="Field14">
          <n-input
            v-model:value="row.Field14"
            :placeholder="$t('model.plugin_data.field14')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field15')" path="Field15">
          <n-input
            v-model:value="row.Field15"
            :placeholder="$t('model.plugin_data.field15')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field16')" path="Field16">
          <n-input
            v-model:value="row.Field16"
            :placeholder="$t('model.plugin_data.field16')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field17')" path="Field17">
          <n-input
            v-model:value="row.Field17"
            :placeholder="$t('model.plugin_data.field17')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field18')" path="Field18">
          <n-input
            v-model:value="row.Field18"
            :placeholder="$t('model.plugin_data.field18')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field19')" path="Field19">
          <n-input
            v-model:value="row.Field19"
            :placeholder="$t('model.plugin_data.field19')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.plugin_data.field20')" path="Field20">
          <n-input
            v-model:value="row.Field20"
            :placeholder="$t('model.plugin_data.field20')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.text_field1')"
          path="TextField1"
        >
          <n-input
            v-model:value="row.TextField1"
            :placeholder="$t('model.plugin_data.text_field1')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.text_field2')"
          path="TextField2"
        >
          <n-input
            v-model:value="row.TextField2"
            :placeholder="$t('model.plugin_data.text_field2')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.text_field3')"
          path="TextField3"
        >
          <n-input
            v-model:value="row.TextField3"
            :placeholder="$t('model.plugin_data.text_field3')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.text_field4')"
          path="TextField4"
        >
          <n-input
            v-model:value="row.TextField4"
            :placeholder="$t('model.plugin_data.text_field4')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.text_field5')"
          path="TextField5"
        >
          <n-input
            v-model:value="row.TextField5"
            :placeholder="$t('model.plugin_data.text_field5')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.index_field1')"
          path="IndexField1"
        >
          <n-input
            v-model:value="row.IndexField1"
            :placeholder="$t('model.plugin_data.index_field1')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.index_field2')"
          path="IndexField2"
        >
          <n-input
            v-model:value="row.IndexField2"
            :placeholder="$t('model.plugin_data.index_field2')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.index_field3')"
          path="IndexField3"
        >
          <n-input
            v-model:value="row.IndexField3"
            :placeholder="$t('model.plugin_data.index_field3')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.index_field4')"
          path="IndexField4"
        >
          <n-input
            v-model:value="row.IndexField4"
            :placeholder="$t('model.plugin_data.index_field4')"
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item
          :label="$t('model.plugin_data.index_field5')"
          path="IndexField5"
        >
          <n-input
            v-model:value="row.IndexField5"
            :placeholder="$t('model.plugin_data.index_field5')"
          /> </n-form-item
      ></n-grid-item>

      <n-grid-item
        ><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
          <n-date-picker
            class="w-full"
            v-model:value="row.CreatedAt"
            type="datetime"
            clearable
          /> </n-form-item
      ></n-grid-item>
      <n-grid-item
        ><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
          <n-date-picker
            class="w-full"
            v-model:value="row.UpdatedAt"
            type="datetime"
            clearable
          /> </n-form-item
      ></n-grid-item> </n-grid
  ></n-form>
</template>
