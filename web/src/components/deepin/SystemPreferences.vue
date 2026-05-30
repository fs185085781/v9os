<script setup>
import { computed, defineAsyncComponent, ref } from "vue";
import { SettingOutlined, SkinOutlined, UserOutlined } from "@vicons/antd";
import { SettingsSuggestOutlined } from "@vicons/material";
import { renderIcon } from "@/util/icon";

const props = defineProps({ winId: { type: String, default: "" } });
const itemIconSize = 28;

const ProfileTab = defineAsyncComponent(
  () => import("@/components/common/component/user/profile/profile.vue"),
);
const PersonalizeTab = defineAsyncComponent(
  () => import("@/components/common/modules/settings/PersonalizeTab.vue"),
);
const SystemSettingsPanel = defineAsyncComponent(
  () => import("@/components/common/modules/settings/SystemSettingsPanel.vue"),
);
const ConfigSettingsPanel = defineAsyncComponent(
  () => import("@/components/common/modules/settings/ConfigSettingsPanel.vue"),
);

const selectedKey = ref("account");
const settingItems = computed(() => {
  const items = [
    {
      key: "account",
      title: $t("common.settings.mine"),
      desc: $user.user?.Name || $user.user?.Username || "",
      icon: renderIcon(UserOutlined, itemIconSize),
    },
    {
      key: "appearance",
      title: $t("common.settings.appearance"),
      desc: $t("backend.layout.settings"),
      icon: renderIcon(SkinOutlined, itemIconSize),
    }
  ]
  if ($user.user?.IsAdmin == 1) {
    items.push({
      key: "system",
      title: $t("common.settings.system"),
      desc: $t("admin.system.websettings"),
      icon: renderIcon(SettingOutlined, itemIconSize),
    },
      {
        key: "config",
        title: $t("common.settings.config"),
        desc: $t("admin.title"),
        icon: renderIcon(SettingsSuggestOutlined, itemIconSize),
      });
  }
  return items;
});

const currentItem = computed(
  () => settingItems.value.find((item) => item.key === selectedKey.value) || settingItems.value[0],
);
const openItem = (key) => {
  selectedKey.value = key;
};
</script>

<template>
  <div class="h-full flex overflow-hidden user-color-fbg user-color-ftext">
    <aside class="w-[218px] shrink-0 user-color-bg-2 px-3 py-5">
      <button v-for="item in settingItems" :key="item.key"
        class="w-full h-12 mb-3 flex items-center gap-4 px-4 text-left user-rounded-2.5"
        :class="selectedKey === item.key ? 'user-color-bg' : 'hover:user-color-control'" @click="openItem(item.key)">
        <component :is="item.icon" class="shrink-0" />
        <span class="truncate text-4 font-600">{{ item.title }}</span>
      </button>
    </aside>

    <main class="min-w-0 flex-1 flex flex-col user-color-fbg">
      <div class="h-13 shrink-0 flex items-center px-6 border-b user-color-line">
        <span class="text-4.5 font-600">{{ currentItem.title }}</span>
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
