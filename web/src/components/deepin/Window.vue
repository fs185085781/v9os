<script setup>
import { computed, ref, nextTick, onMounted ,watch} from 'vue'
import { NIcon } from "naive-ui";
import Vue3DraggableResizable from 'vue3-draggable-resizable'
import 'vue3-draggable-resizable/dist/Vue3DraggableResizable.css'
import GlassLayer from "@/components/common/component/util/GlassLayer.vue";
import IconView from "@/components/common/component/util/IconView.vue";
import { CloseOutlined, FullscreenExitOutlined, FullscreenOutlined, MinusOutlined } from "@vicons/antd";
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
  if (winData.top < 0) {
    winData.top = 0
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
      :class="{ 'hidden': winData.status == 'min', 'animate-[blink_0.1s_ease-in-out_infinite]': !!winData.inBlink }">
      <div class="w-full h-full overflow-hidden user-color-fbg user-color-ftext user-color-border" :class="{
        'user-rounded-4': winData.status != 'max',
        'user-rounded-0': winData.status == 'max',
      }">
        <div @mousedown="tapOnWindow(false)" :class="{ hidden: !(!winData.active || winCommon.inDraggable) }"
          class="absolute mt-12 w-full h-[calc(100%-48px)] z-999999"></div>
        <div class="relative h-10 flex items-center justify-between box-border select-none user-color-fbg border-b border-gray/30"
          :class="{ 'user-rounded-t-4': winData.status != 'max', 'user-rounded-0': winData.status == 'max' }"
          @mousedown="tapOnWindow(true)" @dblclick="winStatusChange(winData.status == 'max' ? 'normal' : 'max')">
          <div class="flex flex-1 items-center gap-2.5 min-w-0 pl-4">
            <IconView v-if="winData.icon" :icon="winData.icon" :size="22" icon-class="w-5.5 h-5.5 object-contain" />
            <span class="text-3.5 font-600 overflow-hidden text-ellipsis whitespace-nowrap user-color-ftext" :class="{ 'select-none': winCommon.inDraggable }" :title="winData.title">
              {{ winData.title }}
            </span>
          </div>
          <div class="h-10 flex flex-row items-center flex-none pr-0">
            <button v-if="!winData.parentId" class="w-10 h-10 flex items-center justify-center hover:bg-[color-mix(in_srgb,var(--user-text-color)_10%,transparent)]"
              @click.stop="winStatusChange('min')">
              <n-icon :component="MinusOutlined" size="16" />
            </button>
            <button v-if="winData.status == 'normal' && !winData.parentId" class="w-10 h-10 flex items-center justify-center hover:bg-[color-mix(in_srgb,var(--user-text-color)_10%,transparent)]"
              @click.stop="winStatusChange('max')">
              <n-icon :component="FullscreenOutlined" size="16" />
            </button>
            <button v-if="winData.status == 'max' && !winData.parentId" class="w-10 h-10 flex items-center justify-center hover:bg-[color-mix(in_srgb,var(--user-text-color)_10%,transparent)]"
              @click.stop="winStatusChange('normal')">
              <n-icon :component="FullscreenExitOutlined" size="16" />
            </button>
            <button class="w-10 h-10 flex items-center justify-center hover:bg-[color-mix(in_srgb,var(--user-text-color)_10%,transparent)]" @click.stop="winStatusChange('close')">
              <n-icon :component="CloseOutlined" size="16" />
            </button>
          </div>
        </div>
        <div class="overflow-auto user-color-fbg user-color-ftext"
          :class="{ 'user-rounded-b-4': winData.status != 'max', 'user-rounded-0': winData.status == 'max', 'select-none': winCommon.inDraggable }"
          :style="{ height: `${winData.height - 48}px` }" @mousedown="tapOnWindow(false)">
        <iframe ref="iframeUi" v-if="winData.iframeUrl" :src="winData.iframeUrl" allow="fullscreen" sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads allow-modals allow-pointer-lock allow-presentation allow-top-navigation allow-top-navigation-by-user-activation" class="w-full h-full"></iframe>
          <component v-else-if="winData.component" :is="winData.component" :data="winData.data" :winId="winData.id" />
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
