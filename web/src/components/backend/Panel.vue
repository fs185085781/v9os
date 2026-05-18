<script setup>
import { defineAsyncComponent, computed, ref } from "vue";
import { useEventBus } from "@/util/event.js";
const Window = defineAsyncComponent(() => import("./Window.vue"));
const Layout = defineAsyncComponent(() => import("./Layout.vue"));
const ws = $wins;
const onlineMode = ref(true);
useEventBus("sw-change-status", (data) => {
  onlineMode.value = data.mode != "offline";
});
</script>
<template>
  <div
    class="w-full h-full overflow-hidden bg-center bg-cover screen-brightness"
  >
    <div class="window-bound z-10 absolute">
      <template v-for="win in ws.windows">
        <Window :data="win" :class="{ hidden: win.close|| win.status == 'min' }" />
      </template>
    </div>
    <Layout />
    <div
      class="fixed left-1/2 -translate-x-100px top-15px text-red-500 text-lg z-9999"
      v-if="!onlineMode"
    >
      当前为离线模式,数据均为缓存,无法真实操作,请熟知
    </div>
  </div>
</template>
<style scoped></style>
