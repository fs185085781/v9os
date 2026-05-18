<script setup>
import { ref, defineAsyncComponent } from "vue";
import { NForm, NFormItem, NInput, NButton, NSpace } from "naive-ui";
import { postData } from "@/util/util";
import { windowsStore } from "@/stores/windows";
import emitter from "@/util/event.js";
let SecurityVerify = null;
const mod = window.__INFACE_MODS__["/src/components/common/inface/user_ee/com/com/SecurityVerify.vue"];
if (mod) {
  SecurityVerify = defineAsyncComponent(
    mod
  )
}
const props = defineProps({ data: {}, winId: { type: String, default: "" } });
const ws = windowsStore();
const { action } = props.data;
const code = ref("");
const loading = ref(false);
async function doVerify() {
  loading.value = true;
  const res = await postData(
    "user",
    "verifyPassword",
    { password: code.value, action },
    "okerr",
  );
  loading.value = false;
  if (res && res.verifyToken) {
    ws.closeWindow(props.winId);
    emitter.emit("profile-verify-done", {
      verifyToken: res.verifyToken,
      action,
    });
  }
}
async function initData(){
  if($user.user.Otp || $user.user.Phone || $user.user.Email){
    $msg.message.error($t("component.user.security_verify.no_security"));
    ws.closeWindow(props.winId);
    emitter.emit("profile-verify-done", {
      verifyToken: "",
      action,
    });
  }
}
initData();
</script>
<template>
  <template v-if="SecurityVerify">
    <SecurityVerify :data="props.data" :winId="props.winId" />
  </template>
  <template v-else>
    <div class="p-4">
      <n-form label-placement="left" label-width="80">
        <n-form-item :label="$t('component.user.security_verify.old_pwd')">
          <n-input v-model:value="code" type="password" :placeholder="$t('component.user.security_verify.old_pwd_placeholder')" @keyup.enter="doVerify" />
        </n-form-item>
      </n-form>
      <n-space justify="end">
        <n-button @click="ws.closeWindow(winId)">{{
          $t("common.all.cancel")
          }}</n-button>
        <n-button type="primary" :loading="loading" :disabled="!code" @click="doVerify">{{ $t("component.user.security_verify.verify")
          }}</n-button>
      </n-space>
    </div>
  </template>

</template>
