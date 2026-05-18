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
const colorOptions = computed(() => [
    {
        label: $t("model.user_settings.color_select_blue"),
        value: "blue"
    },{
        label: $t("model.user_settings.color_select_brown"),
        value: "brown"
    },{
        label: $t("model.user_settings.color_select_cyan"),
        value: "cyan"
    },{
        label: $t("model.user_settings.color_select_deepBlue"),
        value: "deepBlue"
    },{
        label: $t("model.user_settings.color_select_deepPurple"),
        value: "deepPurple"
    },{
        label: $t("model.user_settings.color_select_diy"),
        value: "diy"
    },{
        label: $t("model.user_settings.color_select_gray"),
        value: "gray"
    },{
        label: $t("model.user_settings.color_select_green"),
        value: "green"
    },{
        label: $t("model.user_settings.color_select_orange"),
        value: "orange"
    },{
        label: $t("model.user_settings.color_select_pink"),
        value: "pink"
    },{
        label: $t("model.user_settings.color_select_purple"),
        value: "purple"
    },{
        label: $t("model.user_settings.color_select_red"),
        value: "red"
    },{
        label: $t("model.user_settings.color_select_yellow"),
        value: "yellow"
    },
])
const langOptions = computed(() => [
    {
        label: $t("model.user_settings.lang_select_en"),
        value: "en"
    },{
        label: $t("model.user_settings.lang_select_zh"),
        value: "zh"
    },
])
const themeOptions = computed(() => [
    {
        label: $t("model.user_settings.theme_select_dark"),
        value: "dark"
    },{
        label: $t("model.user_settings.theme_select_light"),
        value: "light"
    },
])
const modeOptions = computed(() => [
    {
        label: $t("model.user_settings.mode_select_backend"),
        value: "backend"
    },{
        label: $t("model.user_settings.mode_select_deepin"),
        value: "deepin"
    },{
        label: $t("model.user_settings.mode_select_macos"),
        value: "macos"
    },{
        label: $t("model.user_settings.mode_select_pad"),
        value: "pad"
    },{
        label: $t("model.user_settings.mode_select_win10"),
        value: "win10"
    },
])
const roundOptions = computed(() => [
    {
        label: $t("model.user_settings.round_select_false"),
        value: "false"
    },{
        label: $t("model.user_settings.round_select_true"),
        value: "true"
    },
])
const default_wallpaper_typeOptions = computed(() => [
    {
        label: $t("model.user_settings.default_wallpaper_type_select_image"),
        value: "image"
    },{
        label: $t("model.user_settings.default_wallpaper_type_select_video"),
        value: "video"
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
  ><n-grid cols="0:1 800:2 1200:3 1600:4 2000:5 2400:6" :x-gap="12"><n-grid-item><n-form-item :label="$t('model.user_settings.color')" path="Color">
      <n-select v-model:value="row.Color" :options="colorOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.color_desc')" path="ColorDesc">
      <n-input v-model:value="row.ColorDesc" :placeholder="$t('model.user_settings.color_desc')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.lang')" path="Lang">
      <n-select v-model:value="row.Lang" :options="langOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.theme')" path="Theme">
      <n-select v-model:value="row.Theme" :options="themeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.mode')" path="Mode">
      <n-select v-model:value="row.Mode" :options="modeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.font')" path="Font">
      <n-input v-model:value="row.Font" :placeholder="$t('model.user_settings.font')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.round')" path="Round">
      <n-select v-model:value="row.Round" :options="roundOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.default_wallpaper')" path="DefaultWallpaper">
      <n-input v-model:value="row.DefaultWallpaper" :placeholder="$t('model.user_settings.default_wallpaper')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.default_wallpaper_type')" path="DefaultWallpaperType">
      <n-select v-model:value="row.DefaultWallpaperType" :options="default_wallpaper_typeOptions" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.dock_apps')" path="DockApps">
      <n-input v-model:value="row.DockApps" :placeholder="$t('model.user_settings.dock_apps')" />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.user_settings.user_id')" path="UserID">
      <n-input v-model:value="row.UserID" :placeholder="$t('model.user_settings.user_id')" />
    </n-form-item></n-grid-item>
    
    <n-grid-item><n-form-item :label="$t('model.common.createdat')" path="CreatedAt">
      <n-date-picker class="w-full" v-model:value="row.CreatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    <n-grid-item><n-form-item :label="$t('model.common.updatedat')" path="UpdatedAt">
      <n-date-picker class="w-full" v-model:value="row.UpdatedAt" type="datetime" clearable />
    </n-form-item></n-grid-item>
    </n-grid></n-form>
</template>