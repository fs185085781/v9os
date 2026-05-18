<script setup>
import { computed, defineAsyncComponent } from "vue";
import "./styles/index.css";
import { useEventBus } from "@/util/event.js";

const Login = defineAsyncComponent(() => import("./Login.vue"));
const Shutdown = defineAsyncComponent(() => import("./Shutdown.vue"));
const Desktop = defineAsyncComponent(() => import("./Desktop.vue"));
const ContextMenu = defineAsyncComponent(() => import("./ContextMenu.vue"));
$user.registerBuiltinApps([
  {
    code: "__settings__",
    name: computed(() => $t("common.settings.title")),
    icon: "/assets/win10/img/settings/settings.png",
    type: "system",
    url: defineAsyncComponent(() => import("./SystemPreferences.vue")),
  },
]);

useEventBus("token-expired", () => {
  $user.user.ID = 0;
});
</script>

<template>
  <Shutdown v-if="$user.system.Shutdown" />
  <Desktop
    v-if="$user.user.ID > 0"
    :style="{ display: $user.system.Shutdown ? 'none' : '' }"
  />
  <ContextMenu v-if="$user.user.ID > 0" />
  <Login v-else />
</template>
