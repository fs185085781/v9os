<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { getWinSize } from "@/util/util.js";
import { useStore } from "@/stores/user.js";
import { openDesktopShortcutWin } from "@/components/common/modules/desktop/shortcut.js";
import IconView from "@/components/common/component/util/IconView.vue";

const store = useStore();
const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});
const searchKeyword = ref("");
const myApps = ref([]);

onMounted(async () => {
  myApps.value = await $user.getMyApps();
});

const filteredApps = computed(() => {
  const kw = searchKeyword.value.trim().toLowerCase();
  if (!kw) return myApps.value;
  return myApps.value.filter((app) => {
    const name = String(app.name?.value || app.name || "");
    return name.toLowerCase().includes(kw) || String(app.code || "").toLowerCase().includes(kw);
  });
});

const appName = (app) => app.name?.value || app.name || app.code;

function close() {
  if (props.winId) {
    $wins.closeWindow(props.winId);
  }
}

function openApp(app) {
  close();
  const sz = getWinSize();
  const options = {
    id: app.windowId || "",
    width: sz.width,
    height: sz.height - 78,
    title: appName(app),
    icon: app.icon,
    data: {},
  };
  if (app.type === "system") {
    options.component = app.url;
  } else if (app.type === "plugin" || app.type === "url") {
    options.iframeUrl = app.url;
  }
  $wins.addWindow(options);
}

function addToDock(app) {
  if (store.addDockApp && app?.code) {
    store.addDockApp(app.code);
  }
}

function addToDesktop(app) {
  openDesktopShortcutWin(app, { title: "添加桌面快捷方式" });
}

const appContext = (app) => ({
  type: "deepin.launchpad.app",
  payload: app,
  actions: {
    addToDock: () => addToDock(app),
    addToDesktop: () => addToDesktop(app),
  },
});

const appMenuResolver = () => [
  {
    key: "addToDock",
    label: () => $t("contextMenu.common.addToDock"),
    group: "main",
    order: 10,
    visible: (ctx) => !$user.getDockApps().includes(ctx.payload.code),
    action: (ctx) => ctx.actions.addToDock(),
  },
  {
    key: "addToDesktop",
    label: () => "添加到桌面",
    group: "main",
    order: 20,
    action: (ctx) => ctx.actions.addToDesktop(),
  },
];

$contextMenu.register("deepin.launchpad.app", appMenuResolver);
onUnmounted(() => {
  $contextMenu.unregister("deepin.launchpad.app", appMenuResolver);
});
</script>

<template>
  <div class="h-full w-full overflow-auto bg-[color-mix(in_srgb,var(--user-bg-color)_72%,transparent)] user-color-ftext" @click.self="close">
    <div class="mx-auto mt-7 w-72 user-rounded-2 bg-[color-mix(in_srgb,var(--user-control-color)_72%,transparent)] px-3 py-1.5 shadow">
      <input
        v-model="searchKeyword"
        class="w-full bg-transparent outline-none text-center text-14px user-color-ftext placeholder:opacity-60"
        placeholder="搜索应用"
        @click.stop
      />
    </div>
    <div class="mx-auto mt-8 grid max-w-260 grid-cols-4 gap-x-8 gap-y-8 px-8 pb-8 sm:grid-cols-6 lg:grid-cols-8">
      <button
        v-for="app in filteredApps"
        :key="app.code"
        class="h-28 min-w-0 flex flex-col items-center justify-start user-rounded-3 p-2 text-center transition hover:bg-[color-mix(in_srgb,var(--user-text-color)_12%,transparent)]"
        @click.stop="openApp(app)"
        v-context-menu="appContext(app)"
      >
        <IconView :icon="app.icon" :alt="appName(app)" :size="56" icon-class="h-14 w-14 object-contain" />
        <span class="mt-2 max-w-full truncate text-13px font-600">{{ appName(app) }}</span>
      </button>
      <div v-if="filteredApps.length === 0" class="col-span-full mt-20 text-center user-color-muted">
        {{ searchKeyword ? "没有匹配的应用" : "暂无可用应用" }}
      </div>
    </div>
  </div>
</template>
