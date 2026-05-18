<script setup>
import { onUnmounted, reactive } from "vue";
const status = reactive({
  onIng: false,
  progress: 0.0,
});
let timer = null;
const powerOn = () => {
  $user.system.Open = false;
  if ($user.system.Wakeup) {
    $user.system.Shutdown = false;
  } else {
    status.onIng = true;
    if (timer) clearInterval(timer);
    timer = setInterval(() => {
      status.progress += 0.5;
      if (status.progress >= 100) {
        clearInterval(timer);
        timer = null;
        $user.system.Shutdown = false;
      }
    }, 15);
  }
  $user.system.Wakeup = false;
};
if ($user.system.Open) {
  powerOn();
}
onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>
<template>
  <div @click="powerOn()" class="w-full h-full bg-black flex-center flex-col">
    <img
      class="w-24 h-24 my-0 mx-auto rounded-full"
      :src="$user.webSettings.Logo"
      alt="img"
    />
    <div
      v-if="!status.onIng"
      class="absolute top-1/2 left-0 right-0"
      m="t-16 sm:t-20 x-auto"
      text="sm gray-200 center"
    >
      {{
        $t("common.all.click", {
          key: $t(
            $user.system.Wakeup ? "common.all.wakeup" : "common.all.poweron",
          ),
        })
      }}
    </div>
    <div
      v-else
      class="absolute top-1/2 left-0 right-0 w-56 h-1 sm:h-1.5 bg-gray-500 user-rounded-1 overflow-hidden"
      m="t-16 sm:t-24 x-auto"
    >
      <span
        class="absolute top-0 bg-white h-full user-rounded-sm"
        :style="{ width: status.progress + '%' }"
      ></span>
    </div>
  </div>
</template>
<style scoped></style>
