<script setup>
import { ref, nextTick, onMounted } from "vue";
import Vue3DraggableResizable from "vue3-draggable-resizable";
import "vue3-draggable-resizable/dist/Vue3DraggableResizable.css";
import emitter from "@/util/event.js";
//读取属性
const props = defineProps({
  data: {},
});
//读取窗体数据
const winData = props.data;
//定义是否可拖拽
const draggable = ref(false);
const resizable = ref(false);
const winCommon = $wins.common
const startEndDrag = (flag) => {
  if (flag) {
    tapOnWindow(false);
    document.body.classList.add("disable-select");
  } else {
    document.body.classList.remove("disable-select");
  }
  winCommon.inDraggable = flag
  winCommon.id = winData.id
  if (winData.left + winData.width < 100) {
    winData.left = 100 - winData.width;
  }
  if (winData.top < 0) {
    winData.top = 0;
  }
  if (winData.top > window.innerHeight - 100) {
    winData.top = window.innerHeight - 100;
  }
  if (winData.left > window.innerWidth - 100) {
    winData.left = window.innerWidth - 100;
  }
};
const winStatusChange = async (status) => {
  $wins.winStatusChange(winData, status, tapOnWindow)
}
const tapOnWindow = (isHeader) => {
  if ($wins.activeWindow(winData)) {
    resizable.value = winData.status != "max";
    draggable.value = winData.status != "max" && isHeader;
  }
};

const deactivated = async () => {
  await nextTick();
  const has = $wins.windows.filter((x) => x.active);
  if (has.length == 0) {
    winData.active = true;
  }
};
const iframeUi = ref(null);
$wins.initPostMessage(iframeUi, winData.id);
onMounted(() => {
  tapOnWindow(false);
});
</script>
<template>
  <Vue3DraggableResizable :init-w="winData.width" :init-h="winData.height" v-model:x="winData.left"
    v-model:y="winData.top" v-model:w="winData.width" v-model:h="winData.height" v-model:active="winData.active"
    :draggable="draggable" :resizable="resizable && !winData.parentId"
    :handles="['tl', 'tm', 'tr', 'ml', 'mr', 'bl', 'bm', 'br']" :min-w="100" :min-h="80"
    @drag-start="startEndDrag(true)" @resize-start="startEndDrag(true)" @drag-end="startEndDrag(false)"
    @resize-end="startEndDrag(false)" @deactivated="deactivated" class="!border-0" :class="{
      'z-11': winData.active,
      'z-10': !winData.active && winCommon.parentId == winData.id,
      'z-9': !winData.active && winCommon.parentId != winData.id,
      hidden: winData.status == 'min',
      'animate-[blink_0.1s_ease-in-out_infinite]': !!winData.inBlink,
    }">
    <div>
      <div @mousedown="tapOnWindow(false)" :class="{ hidden: !(!winData.active || winCommon.inDraggable) }"
        class="mt-6 w-full h-[calc(100%-1.5rem)] absolute z-12"></div>
      <div
        class="user-rounded-t-2.5 window-bar relative h-10 flex items-center justify-between pl-3 pr-0 user-color-fbg user-color-ftext border border-b-0"
        @mousedown="tapOnWindow(true)" @dblclick="winStatusChange(winData.status == 'max' ? 'normal' : 'max')">
        <!-- 左侧：窗口标题 -->
        <span class="font-medium text-sm user-color-ftext select-none truncate flex-1"
          :class="{ 'unselectable': winCommon.inDraggable }" :title="winData.title">
          {{ winData.title }}
        </span>

        <!-- 右侧：窗口控制按钮 -->
        <div class="flex items-center h-full">
          <button v-if="winData.status == 'normal' && !winData.parentId" @click.stop="winStatusChange('max')"
            class="w-10 h-10 flex items-center justify-center user-rounded-1 hover:user-color-bg transition-colors z-13">
            <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" height="14" width="14">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
            </svg>
          </button>
          <button v-if="winData.status == 'max' && !winData.parentId" @click.stop="winStatusChange('normal')"
            class="w-10 h-10 flex items-center justify-center user-rounded-1 hover:user-color-bg transition-colors z-13">
            <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" height="14" width="14">
              <rect x="2" y="6" width="14" height="14" rx="1"></rect>
              <path d="M8 2h12a2 2 0 0 1 2 2v12"></path>
            </svg>
          </button>
          <button @click.stop="winStatusChange('close')"
            class="w-10 h-10 flex items-center justify-center user-rounded-1 hover:user-color-bg transition-colors z-13">
            <svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" height="14" width="14">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
      <div :class="{ 'unselectable': winCommon.inDraggable }"
        class="user-rounded-b-2.5 innner-window w-full overflow-y-auto user-color-fbg user-color-ftext border border-t-0"
        :style="{ height: `${winData.height - 40}px` }" @mousedown="tapOnWindow(false)">
        <iframe ref="iframeUi" v-if="winData.iframeUrl" :src="winData.iframeUrl" allow="fullscreen" sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads allow-modals allow-pointer-lock allow-presentation allow-top-navigation allow-top-navigation-by-user-activation"  class="w-full h-full"></iframe>
        <component v-else-if="winData.component" :is="winData.component" :data="winData.data" :winId="winData.id">
        </component>
      </div>
    </div>
  </Vue3DraggableResizable>
</template>
<style>
.window-bound .vdr-container .vdr-handle {
  display: block !important;
}

.vdr-handle-tm,
.vdr-handle-bm {
  width: 100%;
  left: 0%;
}

.vdr-handle-ml,
.vdr-handle-mr {
  height: 100%;
  top: 0%;
}

.vdr-handle-tl,
.vdr-handle-tr,
.vdr-handle-bl,
.vdr-handle-br {
  z-index: 9;
}

.vdr-handle {
  background: transparent;
  border: 0px;
}

.vdr-container .unselectable {
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
}
</style>
