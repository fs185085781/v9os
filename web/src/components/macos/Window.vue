<script setup>
import { computed, ref, nextTick, onMounted ,watch} from 'vue'
import Vue3DraggableResizable from 'vue3-draggable-resizable'
import 'vue3-draggable-resizable/dist/Vue3DraggableResizable.css'
import GlassLayer from "@/components/common/GlassLayer.vue";
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
  if (winData.left + winData.width < 100) {
    winData.left = 100 - winData.width
  }
  if (winData.top < 24) {
    winData.top = 24
  }
  if (winData.top > window.innerHeight - 100) {
    winData.top = window.innerHeight - 100
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
const updateTheme = (newVal) => {
    const color = newVal == "dark" ? "#ffffff1f":"#0000001f";
    document.documentElement.style.setProperty('--window-box-shadow-color', color);
}
$wins.initPostMessage(iframeUi, winData.id)
onMounted(() => {
  tapOnWindow(false)
  watch(() => $user.settings.Theme,updateTheme )
  updateTheme($user.settings.Theme)
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
      @resize-end="startEndDrag(false)" @deactivated="deactivated" class="!border-0 shadow-[0_0_30px_var(--window-box-shadow-color)]"
      :style="{ zIndex: windowZIndex }"
      :class="{ 'hidden': winData.status == 'min', 'animate-[blink_0.1s_ease-in-out_infinite]': !!winData.inBlink}">
      <div>
        <div @mousedown="tapOnWindow(false)" :class="{ 'hidden': !(!winData.active || winCommon.inDraggable) }"
          class="mt-6 w-full h-[calc(100%-1.5rem)] absolute z-999999">
        </div>
        <div
          class="user-rounded-t-2.5 window-bar relative h-6 text-center user-color-fbg user-color-ftext border-b border-gray/30"
          @mousedown="tapOnWindow(true)" @dblclick="winStatusChange(winData.status == 'max' ? 'normal' : 'max')">
          <div class="traffic-lights flex flex-row absolute left-0 space-x-2 pl-2 mt-1.5"><button
              @click="winStatusChange('close')" class="macos-window-btn bg-red-500 "><svg stroke="currentColor"
                fill="currentColor" stroke-width="0" viewBox="0 0 512 512" height="11" width="11"
                xmlns="http://www.w3.org/2000/svg">
                <path fill="none" stroke-linecap="round" stroke-linejoin="round" stroke-width="32"
                  d="M368 368L144 144m224 0L144 368"></path>
              </svg></button><button v-if="!winData.parentId" @click="winStatusChange('min')"
              class="macos-window-btn bg-yellow-500 "><svg stroke="currentColor" fill="none" stroke-width="2"
                viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round" height="11" width="11"
                xmlns="http://www.w3.org/2000/svg">
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg></button><button v-if="winData.status == 'normal' && !winData.parentId"
              @click="winStatusChange('max')" class="macos-window-btn bg-green-500 "><svg viewBox="0 0 13 13"
                width="6.5" height="6.5" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd"
                stroke-linejoin="round" stroke-miterlimit="2">
                <path d="M9.26 12.03L.006 2.73v9.3H9.26zM2.735.012l9.3 9.3v-9.3h-9.3z"></path>
              </svg></button><button v-if="winData.status == 'max' && !winData.parentId"
              @click="winStatusChange('normal')" class="macos-window-btn bg-green-500 "><svg viewBox="0 0 19 19"
                width="10" height="10" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd"
                stroke-linejoin="round" stroke-miterlimit="2">
                <path d="M18.373 9.23L9.75.606V9.23h8.624zM.6 9.742l8.623 8.624V9.742H.599z"></path>
              </svg></button></div><span :class="{ 'unselectable': winCommon.inDraggable }"
            class="font-semibold user-color-ftext select-none" :title="winData.title">{{
              winData.title.substring(0, 16) }}</span>
        </div>
        <div :class="{ 'unselectable': winCommon.inDraggable }"
          class="user-rounded-b-2.5 innner-window w-full overflow-y-auto user-color-fbg user-color-ftext"
          :style="{ height: `${winData.height - 24}px` }" @mousedown="tapOnWindow(false)">
          <iframe ref="iframeUi" v-if="winData.iframeUrl" :src="winData.iframeUrl" sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads allow-modals allow-pointer-lock allow-presentation allow-top-navigation allow-top-navigation-by-user-activation" class="w-full h-full"></iframe>
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
.vdr-handle-bl {
  z-index: 9;
}

.vdr-handle {
  background: transparent;
  border: 0px;
}

.vdr-container .traffic-lights:hover svg {
  display: block;
}

.vdr-container .traffic-lights svg {
  display: none;
}

.vdr-container .traffic-lights {
  filter: grayscale(1);
}

.vdr-container.active .traffic-lights {
  filter: initial;
}

.vdr-container .unselectable {
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
}
</style>
