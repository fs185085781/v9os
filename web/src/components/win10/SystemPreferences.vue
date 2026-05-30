<script setup>
import { computed, defineAsyncComponent, ref } from "vue";
import { SettingOutlined, SkinOutlined, UserOutlined } from "@vicons/antd";
import { SettingsSuggestOutlined } from "@vicons/material";
import { renderIcon } from "@/util/icon";

const props = defineProps({ winId: { type: String, default: "" } });
const itemIconSize = 44;

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

const activeKey = ref("");
const settingItems = computed(() => {
  const items = [
    {
      key: "account",
      title: $t("common.settings.mine"),
      icon: renderIcon(UserOutlined, itemIconSize),
    },
    {
      key: "appearance",
      title: $t("common.settings.appearance"),
      icon: renderIcon(SkinOutlined, itemIconSize),
    }
  ]
  if ($user.user?.IsAdmin == 1) {
    items.push({
      key: "system",
      title: $t("common.settings.system"),
      icon: renderIcon(SettingOutlined, itemIconSize),
    },
      {
        key: "config",
        title: $t("common.settings.config"),
        icon: renderIcon(SettingsSuggestOutlined, itemIconSize),
      })
  }
  return items
});
</script>

<template>
  <div class="h-full overflow-hidden user-color-fbg user-color-ftext">
    <div v-if="!activeKey" class="h-full overflow-y-auto px-8 pt-8 pb-10">
      <div class="flex flex-wrap" style="column-gap: 12px; row-gap: 28px">
        <div v-for="item in settingItems" :key="item.key" style="width: 228px">
          <button class="max-w-full min-w-0 h-16 flex items-center gap-4 text-left group" style="width: 228px"
            @click="activeKey = item.key">
            <component :is="item.icon" class="shrink-0" />
            <span class="min-w-0 text-[20px] leading-6 text-[#008000] group-hover:underline truncate">
              {{ item.title }}
            </span>
          </button>
        </div>
      </div>
    </div>

    <div v-else class="relative h-full user-color-fbg user-color-ftext">
      <div class="absolute left-3 top-3 z-10">
        <button
          class="w-9 h-9 flex items-center justify-center user-rounded-1 shadow-[0_1px_4px_rgba(0,0,0,0.18)] hover:user-color-control"
          @click="activeKey = ''" :title="$t('common.all.close')">
          <span class="text-6 leading-none">&lsaquo;</span>
        </button>
      </div>
      <div class="h-full overflow-y-auto user-color-fbg user-color-ftext">
        <ProfileTab v-if="activeKey === 'account'" :winId="props.winId" />
        <PersonalizeTab v-else-if="activeKey === 'appearance'" :winId="props.winId" />
        <SystemSettingsPanel v-else-if="activeKey === 'system'" :winId="props.winId" />
        <ConfigSettingsPanel v-else-if="activeKey === 'config'" :winId="props.winId" />
      </div>
    </div>
  </div>
</template>
