<script setup>
import { ref, reactive, computed, onMounted, defineAsyncComponent } from "vue";
import { NForm, NFormItem, NInput, NButton, NSpace, NTag } from "naive-ui";
import { postData } from "@/util/util";
import { windowsStore } from "@/stores/windows";
import SecurityVerifyCom from "./com/SecurityVerify.vue";
import ChangePwdCom from "./com/ChangePwd.vue";
import { useEventBus } from "@/util/event.js";
let Profile = null;
const mod = window.__INFACE_MODS__["/src/components/common/inface/user_ee/com/profile.vue"];
if (mod) {
  Profile = defineAsyncComponent(
    mod
  )
}
const props = defineProps({ winId: { type: String, default: "" } });
const ws = windowsStore();
const userInfo = reactive({ Username: "", Name: "", Avatar: "" });
const form = reactive({ Name: "", Avatar: "" });
const saving = ref(false);
async function loadUser(isFirst) {
  if(!isFirst){
    await $user.loadUser();
  }
  Object.assign(userInfo, $user.user);
  form.Name = userInfo.Name || "";
  form.Avatar = userInfo.Avatar || "";
}
async function saveProfile() {
  const fields = {};
  if (form.Name !== userInfo.Name) fields.name = form.Name;
  if (form.Avatar !== userInfo.Avatar) fields.avatar = form.Avatar;
  if (!Object.keys(fields).length) return;
  saving.value = true;
  const res = await postData("user", "updateProfile", fields, "okerr");
  saving.value = false;
  if (res) {
    await loadUser();
  }
}
function openVerify(action) {
  ws.addWindow(
    {
      width: 380,
      height: 200,
      title: $tc("component.user.security_verify.title"),
      component: SecurityVerifyCom,
      data: { action },
    },
    props.winId,
  );
}
function openChangePwd(verifyToken) {
  ws.addWindow(
    {
      width: 400,
      height: verifyToken ? 250 : 300,
      title: $tc("component.user.profile.change_pwd"),
      component: ChangePwdCom,
      data: { verifyToken },
    },
    props.winId,
  );
}
function handleChangePwd() {
  openVerify("changePassword");
}
if (!Profile) {
  useEventBus("profile-verify-done", (msg) => {
    if (msg.action === "changePassword" && msg.verifyToken) {
      openChangePwd(msg.verifyToken)
    }
  });
  useEventBus("profile-refresh", () => loadUser());
  onMounted(() => loadUser(true));
}
</script>
<template>
  <template v-if="Profile">
    <Profile :winId="props.winId" />
  </template>
  <template v-else>
    <div class="max-w-150 mx-auto py-4">
      <div class="flex items-center mb-6">
        <img v-if="form.Avatar" :src="form.Avatar" class="w-16 h-16 rounded-full mr-4 shrink-0 object-cover" />
        <div v-else
          class="w-16 h-16 rounded-full mr-4 shrink-0 flex items-center justify-center user-color-bg text-white text-2xl font-600">
          {{ userInfo.Username?.charAt(0)?.toUpperCase() }}
        </div>
        <div class="text-lg font-600">{{ userInfo.Username }}</div>
      </div>
      <n-form label-placement="left" label-width="100">
        <n-form-item :label="$t('component.user.profile.avatar_url')">
          <n-input v-model:value="form.Avatar" :placeholder="$t('component.user.profile.avatar_placeholder')" />
        </n-form-item>
        <n-form-item :label="$t('component.user.profile.nickname')">
          <n-input v-model:value="form.Name" />
        </n-form-item>
      </n-form>
      <n-space>
        <n-button type="primary" :loading="saving" @click="saveProfile">{{
          $t("component.user.profile.save_profile")
          }}</n-button>
        <n-button @click="handleChangePwd">{{
          $t("component.user.profile.change_pwd")
          }}</n-button>
      </n-space>
    </div>
  </template>
</template>
