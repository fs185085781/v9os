<script setup>
defineProps({
  icon: {
    type: [String, Function, Object],
    default: "",
  },
  fallback: {
    type: String,
    default: "",
  },
  alt: {
    type: String,
    default: "",
  },
  iconClass: {
    type: [String, Array, Object],
    default: "",
  },
  size: {
    type: Number,
    default: 0,
  },
});
</script>

<template>
  <component v-if="typeof icon === 'function' || (icon && typeof icon === 'object')" :is="icon" :size="size || undefined" :class="iconClass" />
  <img v-else-if="icon || fallback" :src="icon || fallback" :alt="alt" :class="iconClass" />
</template>
