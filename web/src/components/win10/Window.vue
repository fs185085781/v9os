<script setup>
import { computed, ref, nextTick, onMounted } from 'vue'
import Vue3DraggableResizable from 'vue3-draggable-resizable'
import 'vue3-draggable-resizable/dist/Vue3DraggableResizable.css'
import emitter from '@/util/event.js';
import GlassLayer from "@/components/common/component/util/GlassLayer.vue";
import IconView from "@/components/common/component/util/IconView.vue";
//读取属性
const props = defineProps({
  data: {
  }
})
//读取窗体数据
const winData = props.data
//定义是否可拖拽
const draggable = ref(false)
const resizable = ref(false)
const winCommon = $wins.common
const startEndDrag = (flag) => {
  if (flag) {
    tapOnWindow(false)
    document.body.classList.add('disable-select');
  } else {
    document.body.classList.remove('disable-select');
  }
  winCommon.inDraggable = flag
  winCommon.id = winData.id
  if (winData.left + winData.width < 150) {
    winData.left = 150 - winData.width
  }
  if (winData.top < 0) {
    winData.top = 0
  }
  if (winData.top > window.innerHeight - 80) {
    winData.top = window.innerHeight - 80
  }
  if (winData.left > window.innerWidth - 100) {
    winData.left = window.innerWidth - 100
  }
}
const winStatusChange = async (status) => {
  $wins.winStatusChange(winData, status, tapOnWindow)
}
const tapOnWindow = (isHeader) => {
  if ($wins.activeWindow(winData)) {
    resizable.value = winData.status != "max"
    draggable.value = winData.status != "max" && isHeader
  }
}

const deactivated = async () => {
  await nextTick()
  const has = $wins.windows.filter((x) => x.active)
  if (has.length == 0) {
    winData.active = true
  }
}
const iframeUi = ref(null)
$wins.initPostMessage(iframeUi, winData.id)
onMounted(() => {
  tapOnWindow(false)
})
const windowZIndex = computed(() => {
  return winData.zIndex || 12
})
const glassRadiusClass = computed(() => winData.status == "max" ? "user-rounded-0" : "user-rounded-2xl")
</script>
<template>
  <div>
  <Vue3DraggableResizable :init-w="winData.width" :init-h="winData.height" v-model:x="winData.left"
    v-model:y="winData.top" v-model:w="winData.width" v-model:h="winData.height" v-model:active="winData.active"
    :draggable="draggable" :resizable="resizable && !winData.parentId"
    :handles="['tl', 'tm', 'tr', 'ml', 'mr', 'bl', 'bm', 'br']" :min-w="100" :min-h="80"
    @drag-start="startEndDrag(true)" @resize-start="startEndDrag(true)" @drag-end="startEndDrag(false)"
    @resize-end="startEndDrag(false)" @deactivated="deactivated" class="!border-0"
    :style="{ zIndex: windowZIndex }"
    :class="{ 'hidden': winData.status == 'min', 'animate-[blink_0.1s_ease-in-out_infinite]': !!winData.inBlink }">
    <div>
      <div @mousedown="tapOnWindow(false)" :class="{ 'hidden': !(!winData.active || winCommon.inDraggable) }"
        class="mt-8 w-full h-[calc(100%-2rem)] absolute z-999999">
      </div>
      <div
        class="user-rounded-t-2.5 relative h-8 flex items-center justify-between user-color-fbg user-color-ftext border"
        @mousedown="tapOnWindow(true)" @dblclick="winStatusChange(winData.status == 'max' ? 'normal' : 'max')">
        <div class="flex items-center min-w-0 flex-1 pl-2 pr-2">
          <IconView v-if="winData.icon" :icon="winData.icon" :size="16" icon-class="w-4 h-4 object-contain mr-1.5" />
          <span :class="{ 'unselectable': winCommon.inDraggable }"
            class="text-xs user-color-ftext select-none truncate" :title="winData.title">{{
              winData.title }}</span>
        </div>
        <div class="win10-window-controls flex flex-row h-full"><button v-if="!winData.parentId"
            @click="winStatusChange('min')" class="win10-window-btn user-rounded-2"><svg stroke="currentColor" fill="none"
              stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round" height="11" width="11"
              xmlns="http://www.w3.org/2000/svg">
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg></button><button v-if="winData.status == 'normal' && !winData.parentId"
            @click="winStatusChange('max')" class="win10-window-btn user-rounded-2"><svg stroke="currentColor" fill="none"
              stroke-width="1.8" viewBox="0 0 24 24" height="11" width="11" xmlns="http://www.w3.org/2000/svg">
              <rect x="5" y="5" width="14" height="14"></rect>
            </svg></button><button v-if="winData.status == 'max' && !winData.parentId"
            @click="winStatusChange('normal')" class="win10-window-btn user-rounded-2"><svg stroke="currentColor" fill="none"
              stroke-width="1.8" viewBox="0 0 24 24" height="12" width="12" xmlns="http://www.w3.org/2000/svg">
              <rect x="4" y="8" width="12" height="12"></rect>
              <path d="M8 4h12v12"></path>
            </svg></button><button @click="winStatusChange('close')" class="win10-window-btn close user-rounded-2"><svg
              stroke="currentColor" fill="none" stroke-width="1.8" viewBox="0 0 24 24" height="12" width="12"
              xmlns="http://www.w3.org/2000/svg">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg></button></div>
      </div>
      <div :class="{ 'unselectable': winCommon.inDraggable }"
        class="user-rounded-b-2.5 innner-window w-full overflow-y-auto user-color-fbg user-color-ftext border border-t-0"
        :style="{ height: `${winData.height - 32}px` }" @mousedown="tapOnWindow(false)">
        <iframe ref="iframeUi" v-if="winData.iframeUrl" :src="winData.iframeUrl" allow="fullscreen" sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads allow-modals allow-pointer-lock allow-presentation allow-top-navigation allow-top-navigation-by-user-activation" class="w-full h-full"></iframe>
        <component v-else-if="winData.component" :is="winData.component" :data="winData.data" :winId="winData.id">
        </component>
      </div>
    </div>
  </Vue3DraggableResizable>
  <GlassLayer
    :visible="!winData.close && winData.status != 'min'"
    :top="winData.top"
    :left="winData.left"
    :width="winData.width"
    :height="winData.height"
    :content-z-index="windowZIndex"
    :radius-class="glassRadiusClass"
  />
  </div>
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

.win10-window-btn {
  align-items: center;
  color: var(--user-text-color);
  display: flex;
  height: 30px;
  justify-content: center;
  transition: background 0.12s ease, color 0.12s ease;
  width: 40px;
}

.win10-window-btn:hover {
  background: color-mix(in srgb, var(--user-text-color) 12%, transparent);
}

.win10-window-btn.close:hover {
  background: #e81123;
  color: #fff;
}

.vdr-container .unselectable {
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
}
</style>
