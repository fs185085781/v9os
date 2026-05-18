<script setup>
import { computed, defineAsyncComponent, ref } from "vue";
import { SettingOutlined, SkinOutlined } from "@vicons/antd";
import { SettingsSuggestOutlined } from "@vicons/material";
import { renderIcon } from "@/util/icon";

const props = defineProps({ winId: { type: String, default: "" } });
const itemIconSize = 28;

const ProfileTab = defineAsyncComponent(
  () => import("../common/modules/settings/ProfileTab.vue"),
);
const PersonalizeTab = defineAsyncComponent(
  () => import("../common/modules/settings/PersonalizeTab.vue"),
);
const SystemSettingsPanel = defineAsyncComponent(
  () => import("../common/modules/settings/SystemSettingsPanel.vue"),
);
const ConfigSettingsPanel = defineAsyncComponent(
  () => import("../common/modules/settings/ConfigSettingsPanel.vue"),
);

const selectedKey = ref("account");
const accountItem = computed(() => ({
  key: "account",
  title: $user.user?.Username || $t("common.settings.mine"),
  subtitle: $user.user?.Name || $user.user?.Username || "",
  avatar: $user.user?.Avatar || "",
}));
const accountInitial = computed(
  () => (accountItem.value.title || "?").charAt(0).toUpperCase(),
);
const settingItems = computed(() => {
  const items = [
    {
      key: "appearance",
      title: $t("common.settings.appearance"),
      subtitle: $t("backend.layout.settings"),
      icon: renderIcon(SkinOutlined, itemIconSize),
    }
  ]
  if ($user.user?.IsAdmin == 1) {
    items.push({
      key: "system",
      title: $t("common.settings.system"),
      subtitle: $t("admin.system.websettings"),
      icon: renderIcon(SettingOutlined, itemIconSize),
    },
    {
      key: "config",
      title: $t("common.settings.config"),
      subtitle: $t("admin.title"),
      icon: renderIcon(SettingsSuggestOutlined, itemIconSize),
    })
  }
  return items
});

const currentItem = computed(
  () => selectedKey.value === "account"
    ? { title: $t("common.settings.mine") }
    : settingItems.value.find((item) => item.key === selectedKey.value) || settingItems.value[0],
);

const openItem = (key) => {
  selectedKey.value = key;
};
</script>

<template>
  <div class="h-full flex overflow-hidden user-color-fbg user-color-ftext">
    <aside class="w-[232px] shrink-0 overflow-y-auto border-r user-color-line user-color-bg-2 px-3 py-4">
      <div class="h-4"></div>
      <button class="w-full min-w-0 h-17 mb-5 flex items-center gap-3 px-2.5 text-left user-rounded-2"
        :class="selectedKey === 'account' ? 'user-color-control' : 'hover:user-color-control'"
        @click="openItem('account')">
        <img v-if="accountItem.avatar" :src="accountItem.avatar" alt=""
          class="w-12 h-12 shrink-0 object-cover user-rounded-full" />
        <span v-else
          class="w-12 h-12 shrink-0 flex items-center justify-center user-rounded-full user-color-bg text-5 font-700">
          {{ accountInitial }}
        </span>
        <span class="min-w-0 flex-1">
          <span class="block truncate text-[17px] leading-5.5 font-700">{{ accountItem.title }}</span>
          <span class="mt-0.5 block truncate text-[13px] leading-4 user-color-muted">{{ accountItem.subtitle }}</span>
        </span>
      </button>
      <button v-for="item in settingItems" :key="item.key"
        class="w-full h-11 mb-2 flex items-center gap-3 px-2.5 text-left user-rounded-2"
        :class="selectedKey === item.key ? 'user-color-control' : 'hover:user-color-control'"
        @click="openItem(item.key)">
        <component :is="item.icon" class="shrink-0" />
        <span class="min-w-0 truncate text-[15px] leading-5 font-500">{{ item.title }}</span>
      </button>
    </aside>

    <main class="min-w-0 flex-1 flex flex-col user-color-fbg">
      <div class="h-12 shrink-0 flex items-center border-b user-color-line px-5">
        <span class="text-4 font-600">{{ currentItem.title }}</span>
      </div>
      <div class="min-h-0 flex-1 overflow-y-auto user-color-fbg user-color-ftext">
        <ProfileTab v-if="selectedKey === 'account'" :winId="props.winId" />
        <PersonalizeTab v-else-if="selectedKey === 'appearance'" :winId="props.winId" />
        <SystemSettingsPanel v-else-if="selectedKey === 'system'" :winId="props.winId" />
        <ConfigSettingsPanel v-else-if="selectedKey === 'config'" :winId="props.winId" />
      </div>
    </main>
  </div>
</template>
