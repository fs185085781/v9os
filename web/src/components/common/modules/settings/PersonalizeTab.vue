<script setup>
import { reactive, computed, onMounted, ref } from "vue";
import { NSpace, NButton, NSelect, NSwitch, NInput, NSlider } from "naive-ui";
import SingleSelectFile from "@/components/common/component/util/SingleSelectFile.vue";
import { postData,absoluteUrl } from "@/util/util";
const tmdsz = ['win10', 'macos', 'deepin'];
const settings = $user.settings;
const lastMode = settings.Mode;
const colorOptions = computed(() => [
  {
    label: $t("model.user_settings.color_select_green"),
    value: "green",
    color: "#18a058",
  },
  {
    label: $t("model.user_settings.color_select_blue"),
    value: "blue",
    color: "#2080f0",
  },
  {
    label: $t("model.user_settings.color_select_orange"),
    value: "orange",
    color: "#ff9900",
  },
  {
    label: $t("model.user_settings.color_select_purple"),
    value: "purple",
    color: "#722ed1",
  },
  {
    label: $t("model.user_settings.color_select_red"),
    value: "red",
    color: "#d03050",
  },
  {
    label: $t("model.user_settings.color_select_cyan"),
    value: "cyan",
    color: "#0fb9b1",
  },
  {
    label: $t("model.user_settings.color_select_pink"),
    value: "pink",
    color: "#f759ab",
  },
  {
    label: $t("model.user_settings.color_select_yellow"),
    value: "yellow",
    color: "#fadb14",
  },
  {
    label: $t("model.user_settings.color_select_gray"),
    value: "gray",
    color: "#8c8c8c",
  },
  {
    label: $t("model.user_settings.color_select_deepBlue"),
    value: "deepBlue",
    color: "#1d39c4",
  },
  {
    label: $t("model.user_settings.color_select_deepPurple"),
    value: "deepPurple",
    color: "#531dab",
  },
  {
    label: $t("model.user_settings.color_select_brown"),
    value: "brown",
    color: "#ad4e00",
  },
  { label: $t("model.user_settings.color_select_diy"), value: "diy" },
]);

const langOptions = computed(() => [
  { label: $t("model.user_settings.lang_select_zh"), value: "zh" },
  { label: $t("model.user_settings.lang_select_en"), value: "en" },
]);
const modeOptions = computed(() => [
  {
    label: $t("model.user_settings.mode_select_deepin"),
    value: "deepin",
  },
  {
    label: $t("model.user_settings.mode_select_macos"),
    value: "macos",
  },
  {
    label: $t("model.user_settings.mode_select_pad"),
    value: "pad",
  },
  {
    label: $t("model.user_settings.mode_select_win10"),
    value: "win10",
  },
  {
    label: $t("model.user_settings.mode_select_backend"),
    value: "backend",
  },
]);
const fontOptions = ref([{ label: "默认", value: "default" }]);

async function loadFontOptions() {
  try {
    const data = await postData("appstore", "fonts", {}, "");
    if (Array.isArray(data) && data.length) {
      fontOptions.value = data.map((item) => ({
        label: item.label || item.name || item.value,
        value: String(item.value),
      }));
    }
  } catch (error) {
    console.error("load font options failed", error);
  }
}

const roundOptions = computed(() => [
  { label: $t("model.user_settings.round_select_true"), value: "true" },
  { label: $t("model.user_settings.round_select_false"), value: "false" },
]);

const wallpaperTypeOptions = computed(() => [
  {
    label: $t("model.user_settings.default_wallpaper_type_select_image"),
    value: "image",
  },
  {
    label: $t("model.user_settings.default_wallpaper_type_select_video"),
    value: "video",
  },
]);

const diyColor = reactive({ value: settings.ColorDesc || "" });

function setColor(color) {
  $user.setColor(color, true);
}

function setDiyColor() {
  if (diyColor.value) {
    settings.ColorDesc = diyColor.value;
    $user.setColor("diy", true);
  }
}

function setLang(lang) {
  $user.setLang(lang, true);
}

function setMode(mode) {
  $user.setUiMode(mode, true, lastMode);
}

function setTheme(dark) {
  $user.setTheme(dark ? "dark" : "light", true);
}

function setRound(round) {
  $user.setRound(round ? "true" : "false", true);
}

function setFont(font) {
  $user.setFonts(font, true);
}

onMounted(() => {
  loadFontOptions();
});

const isDark = computed(() => settings.Theme === "dark");
const isRound = computed(() => settings.Round === "true");
const transparentInput = reactive({ value: Number(settings.Transparent) || 0 });

function setTransparent(value, save) {
  transparentInput.value = value;
  $user.setTransparent(value, save);
}

// 壁纸
const wallpaperInput = reactive({
  url: settings.DefaultWallpaper || "",
  type: settings.DefaultWallpaperType || "image",
});

function setWallpaper(url, type) {
  $user.setWallpaper(url, type, true);
}

function setCustomWallpaper() {
  if (wallpaperInput.url) {
    setWallpaper(wallpaperInput.url, wallpaperInput.type);
  }
}
function restoreDefaultWallpaper() {
  wallpaperInput.url = "default"
  wallpaperInput.type = "image"
}
</script>

<template>
  <div class="max-w-150 mx-auto py-6 px-8">
    <!-- 外观 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.mode") }}
      </div>
      <n-select :value="settings.Mode" :options="modeOptions" class="w-50" @update:value="setMode" />
    </div>

    <!-- 主题色 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.color") }}
      </div>
      <div class="flex flex-wrap gap-2">
        <div v-for="c in colorOptions.filter((o) => o.color)" :key="c.value" @click="setColor(c.value)"
          class="w-8 h-8 rounded-full cursor-pointer transition-all" :style="{
            backgroundColor: c.color,
            border:
              settings.Color === c.value
                ? '3px solid var(--n-text-color, #333)'
                : '2px solid transparent',
          }" :title="c.label" />
      </div>
      <n-space class="mt-2" align="center">
        <n-input v-model:value="diyColor.value" :placeholder="$t('model.user_settings.color_desc')" class="w-55"
          size="small" />
        <n-button size="small" @click="setDiyColor">{{
          $t("common.all.save")
          }}</n-button>
      </n-space>
    </div>

    <!-- 语言 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.lang") }}
      </div>
      <n-select :value="settings.Lang" :options="langOptions" class="w-50" @update:value="setLang" />
    </div>

    <!-- 深色模式 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.theme") }}
      </div>
      <n-switch :value="isDark" @update:value="setTheme">
        <template #checked>{{
          $t("model.user_settings.theme_select_dark")
          }}</template>
        <template #unchecked>{{
          $t("model.user_settings.theme_select_light")
          }}</template>
      </n-switch>
    </div>

    <!-- 圆角 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.round") }}
      </div>
      <n-switch :value="isRound" @update:value="setRound">
        <template #checked>{{
          $t("model.user_settings.round_select_true")
          }}</template>
        <template #unchecked>{{
          $t("model.user_settings.round_select_false")
          }}</template>
      </n-switch>
    </div>

    <div class="mb-6" v-if="tmdsz.includes(settings.Mode)">
      <div class="font-600 mb-2 user-color-ftext">透明度</div>
      <n-space align="center" class="mt-2">
        <n-slider :value="transparentInput.value" class="w-70" :min="0" :max="100" :step="1"
          @update:value="(value) => setTransparent(value, false)" />
        <span class="w-10 text-center text-sm user-color-ftext">
          {{ transparentInput.value }}%
        </span>
        <n-button size="small" @click="setTransparent(transparentInput.value, true)">{{
          $t("common.all.save")
          }}</n-button>
      </n-space>
    </div>

    <!-- 字体 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.font") }}
      </div>
      <n-select :value="settings.Font" :options="fontOptions" class="w-50" @update:value="setFont" />
    </div>

    <!-- 壁纸 -->
    <div class="mb-6">
      <div class="font-600 mb-2 user-color-ftext">
        {{ $t("model.user_settings.default_wallpaper") }}
      </div>
      <n-space align="center" class="mt-2">
        <n-input v-model:value="wallpaperInput.url" :placeholder="$t('model.user_settings.default_wallpaper')" class="w-75" size="small" />
        <SingleSelectFile size="small" scene="wallpaper" accept=".jpg,.jpeg,.png,.mp4" :val="wallpaperInput.url"
          :onlybtn="true" @change="(p) => wallpaperInput.url = p" />
        <n-select v-model:value="wallpaperInput.type" :options="wallpaperTypeOptions" class="w-25" size="small" />
        <n-button size="small" @click="restoreDefaultWallpaper">{{
          $t("common.settings.restore_default")
          }}</n-button>
        <n-button size="small" @click="setCustomWallpaper">{{
          $t("common.all.save")
          }}</n-button>
      </n-space>
      <div class="mt-2">
        <img v-if="wallpaperInput.type == 'image' && wallpaperInput.url == 'default'"
          class="max-w-[200px] max-h-[100px] object-cover border user-color-border user-rounded-lg" :src="absoluteUrl('/assets/'+$user.settings.Mode+'/img/wallpaper.jpg')" />
        <img v-else-if="wallpaperInput.type == 'image' && wallpaperInput.url"
          class="max-w-[200px] max-h-[100px] object-cover border user-color-border user-rounded-lg" :src="absoluteUrl(wallpaperInput.url)" />
        <video v-if="wallpaperInput.type == 'video' && wallpaperInput.url"
          class="max-w-[200px] max-h-[100px] object-cover border user-color-border user-rounded-lg" :src="absoluteUrl(wallpaperInput.url)" autoplay
          muted loop></video>
      </div>
    </div>
  </div>
</template>
