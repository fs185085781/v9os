import { postData } from "@/util/util";
import { ref, computed, watch } from "vue";

// 语言数据存储
const messages = {};
const currentLocale = ref("");

// 加载本地语言文件
const localePatterns = {
  zh: import.meta.glob("./zh/*.json", { eager: true }),
  en: import.meta.glob("./en/*.json", { eager: true }),
};

// 加载语言消息
export async function loadLocaleMessages(locale) {
  if (currentLocale.value === locale) {
    return;
  }

  const modules = localePatterns[locale];
  if (!modules) {
    return;
  }

  const mergedMessages = {};
  for (const path in modules) {
    Object.assign(mergedMessages, modules[path]);
  }

  // 从服务器获取额外数据
  const res = await postData("lang", "get?lang=" + locale, null, "");
  if (res) {
    Object.assign(mergedMessages, res);
  }

  messages[locale] = mergedMessages;
  currentLocale.value = locale;
}

// 翻译函数
function t(key, values = {}) {
  const localeMessages = messages[currentLocale.value] || {};
  let message = key;

  // 支持嵌套路径访问，如 "admin.title" 或 "model.auth_registry.plugin_code"
  const keys = key.split(".");
  let target = localeMessages;

  for (const k of keys) {
    if (target && typeof target === "object" && k in target) {
      target = target[k];
    } else {
      target = key;
      break;
    }
  }

  if (typeof target === "string") {
    message = target;
  } else {
    message = key;
  }

  // 处理插值
  if (typeof message === "string") {
    for (const [k, v] of Object.entries(values)) {
      message = message.replace(new RegExp(`\\{${k}\\}`), v);
    }
  }

  return message;
}
function tc(key, values = {}) {
  return computed(() => t(key, values));
}

// 暴露给JS使用
window.$t = t;
window.$tc = tc;

// Vue插件：让组件也能使用$t
export default {
  install(app) {
    app.config.globalProperties.$t = t;
    app.config.globalProperties.$tc = tc;
    app.provide("$t", t);
    app.provide("$tc", tc);
    app.provide("$locale", currentLocale);
  },
};

// 初始化
loadLocaleMessages("zh");
