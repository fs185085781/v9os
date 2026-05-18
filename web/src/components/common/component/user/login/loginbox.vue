<template>
  <template v-if="Login">
    <Login />
  </template>
  <template v-else>
    <div class="user-rounded-md mx-auto grid grid-cols-5 w-44 h-8 mt-4 backdrop-blur-2xl" bg="gray-300 opacity-50">
      <input @enter="login" v-model="loginParam.username"
        class="text-sm text-white col-start-1 col-span-4 no-outline bg-transparent px-2" placeholder="用户名" />
    </div>
    <div class="user-rounded-md mx-auto grid grid-cols-5 w-44 h-8 mt-4 backdrop-blur-2xl" bg="gray-300 opacity-50">
      <input @enter="login" v-model="loginParam.password"
        class="text-sm text-white col-start-1 col-span-4 no-outline bg-transparent px-2" type="password"
        placeholder="密码" />
    </div>
    <div class="text-sm mt-2 text-white cursor-pointer text-center bg-transparent p-1" @click="login">
      点击进入
    </div>
  </template>
</template>
<script setup>
import { reactive,defineAsyncComponent } from "vue";
import { NIcon } from "naive-ui";
import { postData } from "@/util/util.js";
import { Password20Regular, Person20Regular } from "@vicons/fluent";
let Login = null;
const mod = window.__INFACE_MODS__["/src/components/common/inface/user_ee/com/login.vue"];
if (mod) {
  Login = defineAsyncComponent(
      mod
  )
}
const loginParam = reactive({
  username: "",
  password: "",
});
const login = async () => {
  const res = await postData("user", "login", loginParam,"okerr");
  if (res) {
    await $user.setToken(res);
    $user.loadUser();
  }
};
</script>
