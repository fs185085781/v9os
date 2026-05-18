import { createDiscreteApi, NInput } from "naive-ui";
import { h, ref } from "vue";
const msg = createDiscreteApi(
  ["message", "dialog", "notification", "loadingBar", "modal"],
  {
    configProviderProps: {},
  },
);
msg.util = {
  confirm: (content, title) => {
    if (!title) {
      title = $t("common.all.tip");
    }
    return new Promise(function (success) {
      msg.dialog.warning({
        title,
        content,
        negativeText: $t("common.all.cancel"),
        positiveText: $t("common.all.confirm"),
        draggable: true,
        closable: false,
        maskClosable: false,
        closeOnEsc: false,
        onPositiveClick: () => {
          success(true);
        },
        onNegativeClick: () => {
          success(false);
        },
      });
    });
  },
  alert: (content, title) => {
    if (!title) {
      title = $t("common.all.tip");
    }
    return new Promise(function (success) {
      msg.dialog.info({
        title,
        content,
        positiveText: $t("common.all.confirm"),
        draggable: true,
        closable: false,
        maskClosable: false,
        closeOnEsc: false,
        onPositiveClick: () => {
          success(true);
        },
      });
    });
  },
  prompt: (content, title) => {
    if (!title) {
      title = $t("common.all.tip");
    }
    return new Promise(function (success) {
      const inputValue = ref("");
      const renderContent = () =>
        h("div", {}, [
          h("div", { style: { marginBottom: "12px" } }, content),
          h(NInput, {
            value: inputValue.value,
            "onUpdate:value": (val) => {
              inputValue.value = val;
            },
          }),
        ]);
      msg.dialog.info({
        title,
        content: renderContent,
        negativeText: $t("common.all.cancel"),
        positiveText: $t("common.all.confirm"),
        draggable: true,
        closable: false,
        maskClosable: false,
        closeOnEsc: false,
        onPositiveClick: () => {
          success(inputValue.value);
        },
        onNegativeClick: () => {
          success("");
        },
      });
    });
  },
};
export function useMsg() {
  return msg;
}
