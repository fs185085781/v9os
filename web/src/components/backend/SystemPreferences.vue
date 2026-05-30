<script setup>
import { computed, defineAsyncComponent, ref } from "vue";
import { UserOutlined, SkinOutlined, SettingOutlined } from "@vicons/antd";
import {SettingsSuggestOutlined} from "@vicons/material";
import { renderIcon } from "@/util/icon";

const props = defineProps({ winId: { type: String, default: "" } });
const navIconSize = 28;

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

const activeKey = ref("account");
const settingItems = computed(() => {
  const items = [
    {
      key: "account",
      icon: renderIcon(UserOutlined, navIconSize),
      title: $t("common.settings.account"),
      desc: $user.user?.Name || $user.user?.Username || "",
    },
    {
      key: "appearance",
      icon: renderIcon(SkinOutlined, navIconSize),
      title: $t("common.settings.appearance"),
      desc: $t("backend.layout.settings"),
    },
  ];
  if ($user.user?.IsAdmin == 1) {
    items.push({
      key: "system",
      icon: renderIcon(SettingOutlined, navIconSize),
      title: $t("common.settings.system"),
      desc: $t("admin.system.websettings"),
    });
    items.push({
      key: "config",
      icon: renderIcon(SettingsSuggestOutlined, navIconSize),
      title: $t("common.settings.config"),
      desc: $t("admin.title"),
    });
  }
  return items;
});

const currentItem = computed(
  () => settingItems.value.find((item) => item.key === activeKey.value) || settingItems.value[0],
);
</script>

<template>
  <div class="h-full flex overflow-hidden user-color-fbg user-color-ftext">
    <aside class="w-60 shrink-0 border-r user-color-line p-3">
      <div class="px-3 py-2 text-5 font-600">{{ $t("common.settings.title") }}</div>
      <div class="mt-3 flex flex-col gap-1">
        <button
          v-for="item in settingItems"
          :key="item.key"
          class="settings-nav-item h-11 flex items-center gap-3 px-3 text-left user-rounded-2 hover:user-color-control"
          :class="{ 'settings-nav-item-active': activeKey === item.key }"
          :aria-current="activeKey === item.key ? 'page' : undefined"
          @click="activeKey = item.key"
        >
          <component :is="item.icon" class="shrink-0" />
          <span class="min-w-0 flex-1 truncate text-3.5">{{ item.title }}</span>
        </button>
      </div>
    </aside>

    <main class="min-w-0 flex-1 flex flex-col">
      <header class="h-14 shrink-0 flex items-center border-b user-color-line px-5">
        <div>
          <div class="text-4 font-600">
            {{ currentItem?.title }}
          </div>
          <div class="mt-0.5 text-3 user-color-muted">
            {{ currentItem?.desc }}
          </div>
        </div>
      </header>
      <div class="min-h-0 flex-1 overflow-y-auto">
        <ProfileTab v-if="activeKey === 'account'" :winId="props.winId" />
        <PersonalizeTab v-else-if="activeKey === 'appearance'" :winId="props.winId" />
        <SystemSettingsPanel v-else-if="activeKey === 'system'" :winId="props.winId" />
        <ConfigSettingsPanel v-else-if="activeKey === 'config'" :winId="props.winId" />
      </div>
    </main>
  </div>
</template>

<style scoped>
.settings-nav-item {
  position: relative;
  width: 100%;
  border: 0;
  color: var(--user-text-1-color);
  transition:
    background-color 0.16s ease,
    color 0.16s ease;
}

.settings-nav-item-active {
  background: var(--user-active-color);
  color: var(--user-primary-color);
  font-weight: 700;
}
</style>
