<script setup>
import { onMounted, onUnmounted, ref, computed } from "vue";
import { getWinSize, elementInMe } from "@/util/util.js";
import { useStore } from "@/stores/user.js";
import emitter, { useEventBus } from "@/util/event.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/component/util/IconView.vue";
const store = useStore();
const launchpadEl = ref();
const searchKeyword = ref("");
const myApps = ref([]); // auth-modules 返回的模块列表
useEventBus("document-click", (el) => {
  if (!elementInMe(launchpadEl._value, el)) {
    emitter.emit("launchpad-pad-show", false);
  }
});
const padClick = (flag) => {
  if (flag) return;
  emitter.emit("launchpad-pad-show", flag);
};
// 从 auth-modules 接口加载模块信息
onMounted(async () => {
  myApps.value = await $user.getMyApps()
});
// 根据用户权限过滤可见的应用
const visibleApps = computed(() => {
  return myApps.value;
});

// 搜索过滤
const filteredApps = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase();
  if (!kw) return visibleApps.value;
  return visibleApps.value.filter(
    (app) =>
      app.name.toLowerCase().includes(kw) ||
      app.code.toLowerCase().includes(kw),
  );
});

const openApp = (app) => {
  emitter.emit("launchpad-pad-show", false);
  const sz = getWinSize();
  const options = {
    id: app.windowId || "",
    width: sz.width,
    height: sz.height,
    title: app.name,
    icon:app.icon,
    data: {},
  };
  if (app.type === "system") {
    options.component = app.url;
  } else if (app.type === "plugin" || app.type === "url") {
    options.iframeUrl = app.url;
  }
  $wins.addWindow(options);
};

const addToDock = (app) => {
  if (store.addDockApp) {
    store.addDockApp(app.code);
  }
};
const addToDesktop = (app) => {
  openDesktopShortcutWin(app, { title: "添加桌面快捷方式" });
};
const appContext = (app) => ({
  type: "macos.launchpad.app",
  payload: app,
  actions: {
    addToDock: () => addToDock(app),
    addToDesktop: () => addToDesktop(app),
  },
});
const appMenuResolver = () => [
  {
    key: "addToDesktop",
    label: () => "添加到桌面",
    group: "main",
    order: 20,
    action: (ctx) => ctx.actions.addToDesktop(),
  },
  {
    key: "addToDock",
    label: () => $t("contextMenu.common.addToDock"),
    group: "main",
    order: 10,
    visible: (ctx) => !$user.getDockApps().includes(ctx.payload.code),
    action: (ctx) => ctx.actions.addToDock(),
  },
];
$contextMenu.register("macos.launchpad.app", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("macos.launchpad.app", appMenuResolver);
});
</script>
<template>
  <div @click="padClick(false)" ref="launchpadEl"
    class="z-30 transform scale-110 w-full h-full fixed overflow-hidden bg-center bg-cover" style="
      background-image: url(img/ui/wallpaper-day.jpg);
      transform: scale(1);
      transition: 0.2s ease-in;
    ">
    <div class="w-full h-full absolute bg-gray-900 bg-opacity-20 backdrop-blur-2xl">
      <div class="mx-auto grid grid-cols-11 h-7 w-64 mt-5 user-rounded-md" bg="gray-200 opacity-10"
        border="1 gray-200 opacity-30">
        <div @click.stop="padClick(true)" class="col-start-1 col-span-1 flex-center">
          <svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 24 24" class="ml-1" color="white"
            height="1em" width="1em" xmlns="http://www.w3.org/2000/svg" style="color: white">
            <path
              d="M10 18a7.952 7.952 0 0 0 4.897-1.688l4.396 4.396 1.414-1.414-4.396-4.396A7.952 7.952 0 0 0 18 10c0-4.411-3.589-8-8-8s-8 3.589-8 8 3.589 8 8 8zm0-14c3.309 0 6 2.691 6 6s-2.691 6-6 6-6-2.691-6-6 2.691-6 6-6z">
            </path>
          </svg>
        </div>
        <input @click.stop="padClick(true)" v-model="searchKeyword"
          class="col-start-2 col-span-10 no-outline bg-transparent px-1 text-sm text-white" placeholder="Search" />
      </div>
      <div class="max-w-launchpad mx-auto mt-8 w-full px-4 sm:px-10 grid" display="grid"
        grid="flow-row cols-4 sm:cols-7">
        <div v-for="app in filteredApps" :key="app.code" class="h-32 sm:h-36 w-full flex-center">
          <div class="h-full w-full flex flex-col cursor-pointer" @click.stop="openApp(app)"
            v-context-menu="appContext(app)">
            <IconView :icon="app.icon" :alt="app.name" :size="80" icon-class="w-20 h-20 mx-auto" />
            <span class="mt-2 mx-auto text-white text-xs sm:text-sm">{{
              app.name
              }}</span>
          </div>
        </div>
        <div v-if="filteredApps.length === 0"
          class="col-span-4 sm:col-span-7 flex-center text-white text-opacity-60 mt-20">
          {{ searchKeyword ? "没有匹配的应用" : "暂无可用应用" }}
        </div>
      </div>
    </div>
  </div>
</template>
