import { reactive } from "vue";
import { defineStore } from "pinia";
import { absoluteUrl, postData } from "@/util/util";

export const webhookStore = defineStore("webhook", () => {
  const scriptMap = reactive({});

  const normalizeSrc = (src, host) => {
    if(src.startsWith("http") || src.startsWith("//")){
      return src;
    }
    return host+src;
  };

  const removeScript = (code) => {
    const record = scriptMap[code];
    if (!record) {
      return;
    }
    record.element?.remove();
    delete scriptMap[code];
  };

  const clear = () => {
    Object.keys(scriptMap).forEach(removeScript);
  };

  const refresh = async () => {
    const data = await postData("plugin", "webhooks", {}, "");
    const hooks = Array.isArray(data) ? data : [];
    const nextMap = {};
    hooks.forEach((item) => {
      if (!item?.code || !item?.src) {
        return;
      }
      nextMap[item.code] = {
        ...item,
        src: normalizeSrc(item.src, absoluteUrl()),
      };
    });

    Object.keys(scriptMap).forEach((code) => {
      const current = scriptMap[code];
      const next = nextMap[code];
      if (!next || next.src !== current.src) {
        removeScript(code);
      }
    });

    Object.values(nextMap).forEach((item) => {
      if (scriptMap[item.code]) {
        return;
      }
      const element = document.createElement("script");
      element.async = true;
      element.dataset.pluginCode = item.code;
      element.src = item.src;
      document.head.appendChild(element);
      scriptMap[item.code] = {
        ...item,
        element,
      };
    });
  };

  window.$webhook = { refresh, clear };
  return window.$webhook;
});
