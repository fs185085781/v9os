<template>
  <n-layout has-sider class="h-full">
    <n-layout-sider bordered collapse-mode="width" :collapsed-width="0" :width="240" show-trigger :collapsed="collapsed"
      @collapse="collapsed = true" @expand="collapsed = false">
      <n-menu v-model:value="activeKey" default-expand-all :collapsed="collapsed" :collapsed-icon-size="28"
        :options="menuOptions" />
    </n-layout-sider>
    <n-layout class="h-full">
      <component :is="pageMap[activeKey]" :winId="winId"></component>
    </n-layout>
  </n-layout>
</template>

<script setup>
import { NIcon, NLayout, NLayoutSider, NMenu } from "naive-ui";
import { h, ref, computed, defineAsyncComponent } from "vue";
import { checkAuth } from "@/directives/auth";
import { adminMenu } from "./admin.js";
const pageMap = {};
const menuData = adminMenu();
for (const value of menuData) {
  for (const child of value.children) {
    let com = null;
    if (child.key.startsWith("inface.") || child.key.startsWith("component.")) {
      const modsMap = child.key.startsWith("inface.") ? window.__INFACE_MODS__ : window.__COMPONENT_MODS__;
      const mod = modsMap[`/src/components/common/${child.key.replaceAll(".", "/")}.vue`]
      if (!mod) {
        continue;
      }
      com = defineAsyncComponent(
        mod
      );
    } else {
      com = defineAsyncComponent(
        () =>
          import(
            `@/components/common/views/${child.pkey}/${child.key}/index.vue`
          ),
      );
    }
    pageMap[child.key] = com;
  }
}
function getMenuLabel(key) {
  let label = null;
  if (key.startsWith("inface.")) {
    label = $t(key.replace("ee.", "ee.tabs."));
  }else if(key.startsWith("component.")){
    label = $t(key.replace("component.", "component.tabs."));
  } else {
    label = $t(`model.${key}.model`) + $t(`admin.table.model`);
  }
  return label;
}
function filterMenuByAuth() {
  const tables = [];
  for (const value of menuData) {
    const table = {
      key: value.key,
      icon: value.icon,
      children: [],
    };
    table.label = $t("admin.system." + table.key);
    for (const child of value.children) {
      if(!pageMap[child.key]){
        continue;
      }
      if (!checkAuth(child.auth || `/api/${child.key}/page`)) {
        continue;
      }
      table.children.push({
        key: child.key,
        icon: child.icon,
        label: getMenuLabel(child.key)
      });
    }
    if (table.children.length > 0) {
      tables.push(table);
    }
  }
  return tables;
}

const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});
const winId = props.winId;
const menuOptions = computed(() => {
  return filterMenuByAuth();
});
const activeKey = ref("");
// 初始化默认选中第一个有权限的菜单项
function initActiveKey() {
  const filtered = menuOptions.value;
  if (
    filtered.length > 0 &&
    filtered[0].children &&
    filtered[0].children.length > 0
  ) {
    activeKey.value = filtered[0].children[0].key;
  }
}
initActiveKey();

const collapsed = ref(false);
</script>
