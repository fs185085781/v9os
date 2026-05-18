<script setup>
import { reactive, computed, onMounted, ref } from "vue";
import { NButton, NInput, NSelect, NSpace, NSwitch } from "naive-ui";
import { postData } from "@/util/util";

const saving = ref(false);
const form = reactive({
  Title: "",
  Subtitle: "",
  Logo: "",
  DefaultPwd: "",
  DefaultColor: "green",
  DefaultLang: "zh",
  DefaultTheme: "light",
  DefaultMode: "macos",
  DefaultFont: "default",
  DefaultRound: "true",
  DefaultWallpaper: "default",
  DefaultWallpaperType: "image",
  Mourning: "false",
  BeianName: "",
  SafeEntry: "",
});

const colorOptions = computed(() => [
  { label: $t("model.user_settings.color_select_green"), value: "green" },
  { label: $t("model.user_settings.color_select_blue"), value: "blue" },
  { label: $t("model.user_settings.color_select_orange"), value: "orange" },
  { label: $t("model.user_settings.color_select_purple"), value: "purple" },
  { label: $t("model.user_settings.color_select_red"), value: "red" },
  { label: $t("model.user_settings.color_select_cyan"), value: "cyan" },
  { label: $t("model.user_settings.color_select_pink"), value: "pink" },
  { label: $t("model.user_settings.color_select_yellow"), value: "yellow" },
  { label: $t("model.user_settings.color_select_gray"), value: "gray" },
  { label: $t("model.user_settings.color_select_deepBlue"), value: "deepBlue" },
  { label: $t("model.user_settings.color_select_deepPurple"), value: "deepPurple" },
  { label: $t("model.user_settings.color_select_brown"), value: "brown" },
]);
const langOptions = computed(() => [
  { label: $t("model.user_settings.lang_select_zh"), value: "zh" },
  { label: $t("model.user_settings.lang_select_en"), value: "en" },
]);
const themeOptions = computed(() => [
  { label: $t("model.user_settings.theme_select_light"), value: "light" },
  { label: $t("model.user_settings.theme_select_dark"), value: "dark" },
]);
const modeOptions = computed(() => [
  { label: $t("model.user_settings.mode_select_macos"), value: "macos" },
  { label: $t("model.user_settings.mode_select_win10"), value: "win10" },
  { label: $t("model.user_settings.mode_select_deepin"), value: "deepin" },
  { label: $t("model.user_settings.mode_select_backend"), value: "backend" },
]);
const wallpaperTypeOptions = computed(() => [
  { label: $t("model.user_settings.default_wallpaper_type_select_image"), value: "image" },
  { label: $t("model.user_settings.default_wallpaper_type_select_video"), value: "video" },
]);
const isRound = computed({
  get: () => form.DefaultRound === "true",
  set: (value) => {
    form.DefaultRound = value ? "true" : "false";
  },
});
const isMourning = computed({
  get: () => form.Mourning === "true",
  set: (value) => {
    setMourning(value);
  },
});

function applyWebSettings(data) {
  Object.keys(form).forEach((key) => {
    if (data?.[key] != null) form[key] = data[key];
  });
}

async function loadWebSettings(isFirst) {
  if(!isFirst){
    await $user.loadUser();
  }
  applyWebSettings($user.webSettings.value || {});
}

async function saveWebSettings() {
  saving.value = true;
  const ok = await postData("system", "settingsSave", { ...form }, "okerr");
  saving.value = false;
  if (ok) await loadWebSettings();
}

function setMourning(value) {
  form.Mourning = value ? "true" : "false";
  $user.setMourning(form.Mourning);
}

onMounted(() => {
  loadWebSettings(true);
});
</script>

<template>
  <div class="max-w-150 mx-auto py-6 px-8">
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.site_title") }}</div>
      <n-input v-model:value="form.Title" :placeholder="$t('common.settings.site_title')" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.site_subtitle") }}</div>
      <n-input v-model:value="form.Subtitle" :placeholder="$t('common.settings.site_subtitle')" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.site_logo") }}</div>
      <n-input v-model:value="form.Logo" :placeholder="$t('common.settings.site_logo')" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.default_pwd") }}</div>
      <n-input v-model:value="form.DefaultPwd" type="password" show-password-on="click" :placeholder="$t('common.settings.default_pwd')" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.lang") }}</div>
      <n-select v-model:value="form.DefaultLang" :options="langOptions" class="w-50" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.mode") }}</div>
      <n-select v-model:value="form.DefaultMode" :options="modeOptions" class="w-50" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.theme") }}</div>
      <n-select v-model:value="form.DefaultTheme" :options="themeOptions" class="w-50" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.color") }}</div>
      <n-select v-model:value="form.DefaultColor" :options="colorOptions" class="w-50" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.font") }}</div>
      <n-input v-model:value="form.DefaultFont" :placeholder="$t('model.user_settings.font')" class="w-50" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.round") }}</div>
      <n-switch v-model:value="isRound" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.mourning") }}</div>
      <n-switch v-model:value="isMourning" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("model.user_settings.default_wallpaper") }}</div>
      <n-space align="center">
        <n-input v-model:value="form.DefaultWallpaper" :placeholder="$t('model.user_settings.default_wallpaper')" class="w-75" />
        <n-select v-model:value="form.DefaultWallpaperType" :options="wallpaperTypeOptions" class="w-25" />
      </n-space>
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.beian_name") }}</div>
      <n-input v-model:value="form.BeianName" :placeholder="$t('common.settings.beian_name')" />
    </div>

    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">{{ $t("common.settings.safe_entry") }}</div>
      <n-input v-model:value="form.SafeEntry" :placeholder="$t('common.settings.safe_entry')" />
    </div>

    <div class="pt-3 border-t user-color-line flex justify-end">
      <n-button type="primary" :loading="saving" @click="saveWebSettings">
        {{ $t("common.all.save") }}
      </n-button>
    </div>
  </div>
</template>
