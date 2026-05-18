const contextMap = new WeakMap();

export const getContextByElement = (target) => {
  let el = target;
  while (el) {
    if (contextMap.has(el)) {
      const value = contextMap.get(el);
      if (typeof value === "function") {
        return value();
      }
      if (value && typeof value === "object" && "value" in value) {
        return value.value;
      }
      return value;
    }
    el = el.parentElement;
  }
  return null;
};

export default {
  mounted(el, binding) {
    contextMap.set(el, binding.value);
  },
  updated(el, binding) {
    contextMap.set(el, binding.value);
  },
  unmounted(el) {
    contextMap.delete(el);
  },
};
