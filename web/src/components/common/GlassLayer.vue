<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

const props = defineProps({
  visible: { type: Boolean, default: true },
  target: { type: Object, default: null },
  top: { type: [Number, String], default: null },
  left: { type: [Number, String], default: null },
  width: { type: [Number, String], default: null },
  height: { type: [Number, String], default: null },
  zIndex: { type: [Number, String], default: 1 },
  contentZIndex: { type: [Number, String], default: null },
  radiusClass: { type: String, default: "user-rounded-2xl" },
  fixed: { type: Boolean, default: false },
});

const targetRect = ref(null);
let resizeObserver = null;

const numberToPx = (value) => (typeof value === "number" ? `${value}px` : value);
const getTargetElement = () => props.target?.value || props.target || null;

const updateTargetRect = () => {
  const el = getTargetElement();
  if (!el) {
    targetRect.value = null;
    return;
  }
  const rect = el.getBoundingClientRect();
  targetRect.value = {
    top: rect.top,
    left: rect.left,
    width: rect.width,
    height: rect.height,
  };
};

const disconnectObserver = () => {
  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }
};

const connectObserver = async () => {
  disconnectObserver();
  if (!props.visible || !props.target) return;
  await nextTick();
  const el = getTargetElement();
  if (!el) return;
  updateTargetRect();
  resizeObserver = new ResizeObserver(updateTargetRect);
  resizeObserver.observe(el);
};

watch(() => [props.visible, props.target], connectObserver, { immediate: true });

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      nextTick(updateTargetRect);
      window.addEventListener("scroll", updateTargetRect, true);
      window.addEventListener("resize", updateTargetRect);
    } else {
      window.removeEventListener("scroll", updateTargetRect, true);
      window.removeEventListener("resize", updateTargetRect);
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  disconnectObserver();
  window.removeEventListener("scroll", updateTargetRect, true);
  window.removeEventListener("resize", updateTargetRect);
});

const layerStyle = computed(() => {
  const rect = targetRect.value;
  const zIndex =
    props.contentZIndex == null ? props.zIndex : Number(props.contentZIndex) - 1;
  return {
    position: props.fixed || props.target ? "fixed" : "absolute",
    zIndex,
    top: numberToPx(props.top ?? rect?.top ?? 0),
    left: numberToPx(props.left ?? rect?.left ?? 0),
    width: numberToPx(props.width ?? rect?.width ?? 0),
    height: numberToPx(props.height ?? rect?.height ?? 0),
  };
});
</script>

<template>
  <div
    v-if="visible"
    class="v9os-glass-layer pointer-events-none"
    :class="radiusClass"
    :style="layerStyle"
  />
</template>
