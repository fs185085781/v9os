<script setup>
import { reactive, ref } from "vue";
import { NForm, NFormItem, NInput, NButton, NSpace } from "naive-ui";
import { postData } from "@/util/util";
import { windowsStore } from "@/stores/windows";
import emitter from "@/util/event.js";

const props = defineProps({ data: {}, winId: { type: String, default: "" } });
const ws = windowsStore();
const { verifyToken } = props.data;
const form = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});
const loading = ref(false);

async function submit() {
  if (!form.newPassword || form.newPassword !== form.confirmPassword) return;
  if (!verifyToken) return;
  loading.value = true;
  const res = await postData(
    "user",
    "changePasswordByToken",
    {
      verifyToken,
      newPassword: form.newPassword,
    },
    "okerr",
  );
  loading.value = false;
  if (res) {
    ws.closeWindow(props.winId);
    emitter.emit("profile-refresh");
  }
}
</script>
<template>
  <div class="p-4">
    <n-form label-placement="left" label-width="90">
      <n-form-item v-if="!verifyToken" :label="$t('component.user.security_verify.old_pwd')">
        <n-input v-model:value="form.oldPassword" type="password" />
      </n-form-item>
      <n-form-item :label="$t('component.user.profile.new_pwd')">
        <n-input v-model:value="form.newPassword" type="password" />
      </n-form-item>
      <n-form-item :label="$t('component.user.profile.confirm_pwd')">
        <n-input v-model:value="form.confirmPassword" type="password" />
      </n-form-item>
    </n-form>
    <n-space justify="end">
      <n-button @click="ws.closeWindow(winId)">{{
        $t("common.all.cancel")
      }}</n-button>
      <n-button type="primary" :loading="loading" @click="submit">{{
        $t("component.user.profile.confirm_change")
      }}</n-button>
    </n-space>
  </div>
</template>
