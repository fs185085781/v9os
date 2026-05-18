<template>
  <n-select
    v-bind="$attrs"
    :value="innerValue"
    multiple
    @update:value="updateValue"
  >
    <template v-for="(_, name) in slots" #[name]="slotProps">
      <slot :name="name" v-bind="slotProps || {}" />
    </template>
  </n-select>
</template>

<script setup>
import { computed, useSlots } from "vue";
import { NSelect } from "naive-ui";

const props = defineProps({
  value: [String, Array],
  modelValue: [String, Array],
  separator: { type: String, default: "," },
});

const emit = defineEmits(["update:value", "update:modelValue"]);
const slots = useSlots();

// 获取当前值
const currentValue = computed(() =>
  props.value !== undefined ? props.value : props.modelValue,
);

// 字符串转数组
const innerValue = computed(() => {
  const val = currentValue.value;

  if (!val && val !== 0) return [];
  if (Array.isArray(val)) return val.filter(Boolean);

  return String(val)
    .split(props.separator)
    .map((item) => item.trim())
    .filter(Boolean);
});

// 数组转字符串
const updateValue = (val) => {
  const result = val?.length ? val.filter(Boolean).join(props.separator) : "";

  emit("update:value", result);
  emit("update:modelValue", result);
};
</script>
