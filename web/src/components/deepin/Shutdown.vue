<script setup>
import { onUnmounted, reactive } from "vue";

const status = reactive({
  onIng: false,
  progress: 0,
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
  <div class="fixed z-9999 w-full h-full overflow-hidden flex flex-col items-center justify-center bg-gradient-to-br from-[rgb(19,16,55)] to-[rgb(7,7,34)] text-gray-100" @click="powerOn">
    <img
      class="relative z-1 w-24 h-24 object-cover user-rounded-full"
      :src="$user.webSettings.Logo"
      alt=""
    />
    <div v-if="!status.onIng" class="relative z-1 mt-8 text-3.5">
      {{
        $t("common.all.click", {
          key: $t($user.system.Wakeup ? "common.all.wakeup" : "common.all.poweron"),
        })
      }}
    </div>
    <div v-else class="relative z-1 mt-10 w-56 h-1.5 overflow-hidden bg-[rgba(180,190,220,0.35)] user-rounded-full">
      <span class="block h-full bg-white user-rounded-full" :style="{ width: status.progress + '%' }"></span>
    </div>
  </div>
</template>
