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
const app_typeOptions = computed(() => [
    {
        label: $t("model.desktop_app.app_type_select_iframe"),
        value: "iframe"
    },{
        label: $t("model.desktop_app.app_type_select_plugin"),
        value: "plugin"
    },{
        label: $t("model.desktop_app.app_type_select_system"),
        value: "system"
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.desktop_app.user_id')" path="UserID">
      <n-input v-model:value="row.UserID" :placeholder="$t('model.desktop_app.user_id')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.icon')" path="Icon">
      <n-input v-model:value="row.Icon" :placeholder="$t('model.desktop_app.icon')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.title')" path="Title">
      <n-input v-model:value="row.Title" :placeholder="$t('model.desktop_app.title')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.app_type')" path="AppType">
      <n-select v-model:value="row.AppType" :options="app_typeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.code')" path="Code">
      <n-input v-model:value="row.Code" :placeholder="$t('model.desktop_app.code')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.url')" path="Url">
      <n-input v-model:value="row.Url" :placeholder="$t('model.desktop_app.url')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.desktop_app.sort')" path="Sort">
      <n-input v-model:value="row.Sort" :placeholder="$t('model.desktop_app.sort')" />
    </n-form-item></n-grid-item>
    
    <n-grid-item><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
      <n-date-picker class="w-full" v-model:value="row.CreatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
      <n-date-picker class="w-full" v-model:value="row.UpdatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    </n-grid></n-form>
</template>