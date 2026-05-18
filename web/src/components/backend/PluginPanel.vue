<script setup>
import { ref } from 'vue';
import { uuid } from "@/util/util"
const props = defineProps({
  data: {},
});
const plugin = props.data;
const iframeUi = ref(null)
$wins.initPostMessage(iframeUi, uuid())
</script>
<template>
    <template v-if="plugin.type == 'system'">
        <component :is="plugin.url"/>
    </template>
    <template v-else-if="plugin.type == 'plugin'">
        <iframe ref="iframeUi" :src="plugin.url" frameborder="0" width="100%" height="100%"></iframe>
    </template>
</template>
