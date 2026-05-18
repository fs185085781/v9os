<script setup>
import { ref, watch } from "vue";
import { NBadge, NButton, NEmpty, NIcon, NPopover, NScrollbar, NTag } from "naive-ui";
import { MessageOutlined } from "@vicons/antd";
import { useChatNotifyCenter } from "./notifyCenter";

const props = defineProps({
  placement: {
    type: String,
    default: "bottom-end",
  },
  triggerClass: {
    type: String,
    default: "inline-flex h-6 cursor-pointer flex-row items-center user-rounded-1 hover:user-color-bg p-1",
  },
  iconSize: {
    type: Number,
    default: 18,
  },
  iconRender: {
    type: Function,
    default: null,
  },
  panelClass: {
    type: String,
    default: "w-86 max-w-[calc(100vw-16px)] user-color-ftext user-color-fbg",
  },
});

const chatNotify = useChatNotifyCenter();
const show = ref(false);

watch(chatNotify.unreadCount, (count) => {
  if (count <= 0) show.value = false;
});

function togglePanel() {
  if (chatNotify.unreadCount.value <= 0) {
    show.value = false;
    chatNotify.openChatCenter();
    return;
  }
  show.value = !show.value;
}

function openRow(row) {
  show.value = false;
  chatNotify.openChatFromMessage(row);
}

function formatTime(time) {
  if (!time) return "";
  return new Date(time).format("MM-dd HH:mm");
}
</script>

<template>
  <n-popover v-model:show="show" trigger="manual" :placement="props.placement" :show-arrow="false" style="padding: 0;">
    <template #trigger>
      <button
        type="button"
        :class="props.triggerClass"
        @click="togglePanel"
      >
        <n-badge :value="chatNotify.unreadCount.value" :max="99" :show="chatNotify.unreadCount.value > 0">
          <component v-if="props.iconRender" :is="props.iconRender" />
          <n-icon v-else :component="MessageOutlined" :size="props.iconSize" />
        </n-badge>
      </button>
    </template>
    <div :class="props.panelClass">
      <div class="h-11 flex items-center justify-between border-b user-color-border px-3">
        <div class="font-600">消息</div>
        <n-button text size="small" type="primary" @click="chatNotify.refresh">刷新</n-button>
      </div>
      <n-scrollbar class="max-h-108">
        <button
          v-for="row in chatNotify.rows.value"
          :key="row.id"
          class="w-full min-h-15 flex items-start gap-2.5 border-0 border-b user-color-border bg-transparent px-3 py-2.5 text-left cursor-pointer hover:user-color-bg"
          @click="openRow(row)"
        >
          <div class="mt-0.5 h-8.5 w-8.5 flex shrink-0 items-center justify-center user-rounded-full bg-[var(--user-primary-color)] text-[var(--user-primary-text-color)]">
            <n-icon :component="MessageOutlined" size="17" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-13px font-600">{{ row.title }}</span>
              <n-tag v-if="row.unread" type="error" size="small" round>未读</n-tag>
            </div>
            <div class="mt-0.5 line-clamp-2 break-words text-12px user-color-muted">{{ row.subtitle }}</div>
          </div>
          <div class="shrink-0 text-11px user-color-muted">{{ formatTime(row.time) }}</div>
        </button>
        <n-empty v-if="chatNotify.rows.value.length === 0" description="暂无消息" class="my-10" />
      </n-scrollbar>
    </div>
  </n-popover>
</template>
