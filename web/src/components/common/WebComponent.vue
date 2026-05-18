<template>
  <iframe
    ref="iframeUi"
    v-if="component.type == 'iframe'"
    :src="component.content"
    sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox allow-downloads allow-modals allow-pointer-lock allow-presentation allow-top-navigation allow-top-navigation-by-user-activation"
    frameborder="0"
    style="background-color: transparent"
    :style="component.css"
  ></iframe>
  <component
    v-else-if="component.type == 'vue'"
    :is="component.content"
    :style="component.css"
    :winId="props.winId"
  ></component>
  <div
    v-else-if="component.type == 'html'"
    v-html="component.content"
    :style="component.css"
  ></div>
  <div
    v-else-if="component.type == 'webcom'"
    v-html="component.content"
    :style="component.css"
  ></div>
</template>
<script setup>
import { useStore } from "@/stores/user.js";
import { computed, onMounted, ref, defineAsyncComponent } from "vue";
import { postData, getApiHost, uuid } from "@/util/util.js";
import DOMPurify from "dompurify";
const componentModules = import.meta.glob("./component/**/*.vue");
const user = useStore();
const props = defineProps({
  com: {
    type: String,
    default: "",
  },
  winId: {
    type: String,
    default: "",
  },
});
const iframeUi = ref(null);
const component = ref({ type: "", css: "" });
onMounted(async () => {
  if (!props.com) {
    return;
  }
  let data = await postData("component", props.com, null, "");
  if (!data) {
    data = {
      ComType: "vue",
      key: "error/error",
    };
  }
  const ah = await getApiHost();
  if (data.ComType == "iframe") {
    const winId = props.winId || uuid();
    component.value.content = computed(() => {
      let prefix = ah;
      if (data.ComContent.startsWith("http")) {
        prefix = "";
      }
      let wh = "?";
      if ((prefix + data.ComContent).includes("?")) {
        wh = "&";
      }
      return `${prefix}${data.ComContent}${wh}color=${user.settings.Color}&lang=${user.settings.Lang}&round=${user.settings.Round}&theme=${user.settings.Theme}&uimode=${user.settings.Mode}&fonts=${user.settings.Font}`;
    });
    $wins.initPostMessage(iframeUi, winId);
  } else if (data.ComType == "vue") {
    if (!data.key) {
      data.key = props.com;
    }
    component.value.content = defineAsyncComponent(
      componentModules[`./component/${data.key}.vue`],
    );
  } else if (data.ComType == "html") {
    component.value.content = DOMPurify.sanitize(data.ComContent);
  } else if (data.ComType == "webcom") {
    const name = props.com.replace("/", "-");
    const zj = customElements.get(name);
    if (!zj) {
      class WebCom extends HTMLElement {
        constructor() {
          super();
          var shadow = this.attachShadow({ mode: "closed" });
          eval(data.ComContent);
        }
      }
      customElements.define(name, WebCom);
    }
    component.value.content = `<${name}></${name}>`;
  }
  component.value.type = data.ComType;
  component.value.css = data.ComCss;
});
</script>
