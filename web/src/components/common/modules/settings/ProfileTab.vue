<script setup>
import WebComponent from "../../../common/WebComponent.vue";
const props = defineProps({
  winId: {
    type: String,
    default: "",
  },
});
</script>

<template>
  <WebComponent com="user/profile/profile" :winId="props.winId" />
</template>
