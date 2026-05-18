<template>
  <iframe
    ref="iframeUi"
    :src="props.url"
    frameborder="0"
    style="width: 100%; height: 100%"
  ></iframe>
</template>
<script setup>
import { ref } from "vue";
const props = defineProps({
  url: { type: String, default: null },
});
import { useEventBus } from "@/util/event.js";
const iframeUi = ref(null);
$wins.initPostMessage(iframeUi, "expand");
useEventBus("expand-change", (msg) => {
  if (window != window.parent) {
    window.parent.postMessage(msg, "*");
  } else {
  }
});
</script>
