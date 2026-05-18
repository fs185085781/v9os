<script setup>
import { defineAsyncComponent } from "vue";
import { useEventBus } from "@/util/event.js";
const Login = defineAsyncComponent(() => import("./Login.vue"));
const Panel = defineAsyncComponent(() => import("./Panel.vue"));
const ContextMenu = defineAsyncComponent(() => import("./ContextMenu.vue"));
useEventBus("token-expired", () => {
  $user.user.ID = 0;
});
</script>
<template>
  <Panel
    v-if="$user.user.ID > 0"
    :style="{ display: $user.system.Shutdown ? 'none' : '' }"
  />
  <ContextMenu v-if="$user.user.ID > 0" />
  <Login v-else />
</template>
