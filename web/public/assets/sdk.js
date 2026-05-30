(function () {
  let src = document.currentScript.src;
  let sz = src.split("/");
  sz.length -= 2;
  const host = sz.join("/");
  document.writeln(`<script src="${host}/assets/core.js"></script>`);
  if (src.includes("naive=true")) {
    document.writeln(
      `<script src="${host}/assets/expand/vue.global.prod.js"></script>`,
    );
    document.writeln(
      `<script src="${host}/assets/expand/index.prod.js"></script>`,
    );
  }
  if(src.includes("jquery=true")){
    document.writeln(
      `<script src="${host}/assets/expand/jquery.min.js"></script>`,
    );
    document.writeln(
      `<script src="${host}/assets/expand/expand.js"></script>`,
    );
  }
  document.writeln(`<style>::-webkit-scrollbar {width: 8px;height: 8px;}::-webkit-scrollbar-track {background: transparent;}
    ::-webkit-scrollbar-thumb {background: #bbbbbb40;border-radius: 6px;}::-webkit-scrollbar-thumb:hover {background: #bbbbbb40;
}* {scrollbar-width: thin;scrollbar-color: #bbbbbb40 transparent;}</style>`);
  const timeout = 60; //方法调用超时为60秒
  const invokeMap = {};
  const eventMap = {};
  const contextMenuActionMap = {};
  const hostEventSubscribed = {};
  let contextMenuOpened = false;
  const postToParent = (action, data) => {
    parent.postMessage(
      {
        __v9os: true,
        version: 1,
        channel: "plugin",
        action,
        data,
      },
      "*",
    );
  };
  const jsonClone = (data) => {
    if (data === undefined) return undefined;
    try {
      return JSON.parse(JSON.stringify(data));
    } catch (e) {
      return data;
    }
  };
  const colorMap = {
    green: ["#36ad6a", "#18a058", "#0c7a43"],
    blue: ["#4098fc", "#2080f0", "#1060c9"],
    orange: ["#ffad33", "#ff9900", "#f29100"],
    purple: ["#9254de", "#722ed1", "#ab7ae0"],
    red: ["#f56c6c", "#d03050", "#c45656"],
    cyan: ["#2de0c9", "#0fb9b1", "#0ea5a5"],
    pink: ["#ff85c0", "#f759ab", "#f5317f"],
    yellow: ["#ffec3d", "#fadb14", "#d4b106"],
    gray: ["#a6a6a6", "#8c8c8c", "#737373"],
    deepBlue: ["#2540e9", "#1d39c4", "#10239c"],
    deepPurple: ["#693ac9", "#531dab", "#3c1380"],
    brown: ["#d46b08", "#ad4e00", "#873b00"],
  };
  const clamp = (value, min, max) => Math.max(min, Math.min(max, value));
  const alphaHex = (alpha) =>
    Math.round(clamp(alpha, 0, 1) * 255)
      .toString(16)
      .padStart(2, "0");
  const parseColor = (color) => {
    let value = String(color || "").trim();
    if (/^#[0-9a-f]{3}$/i.test(value)) {
      value = `#${value
        .slice(1)
        .split("")
        .map((item) => item + item)
        .join("")}`;
    }
    if (/^#[0-9a-f]{6}$/i.test(value)) {
      const num = parseInt(value.slice(1), 16);
      return {
        r: (num >> 16) & 255,
        g: (num >> 8) & 255,
        b: num & 255,
      };
    }
    const match = value.match(
      /rgba?\(\s*([\d.]+)\s*,\s*([\d.]+)\s*,\s*([\d.]+)/i,
    );
    if (match) {
      return {
        r: clamp(Number(match[1]), 0, 255),
        g: clamp(Number(match[2]), 0, 255),
        b: clamp(Number(match[3]), 0, 255),
      };
    }
    return { r: 32, g: 128, b: 240 };
  };
  const rgba = (color, alpha) => {
    const rgb = parseColor(color);
    return `rgba(${Math.round(rgb.r)}, ${Math.round(rgb.g)}, ${Math.round(rgb.b)}, ${clamp(alpha, 0, 1).toFixed(2)})`;
  };
  const luminance = (color) => {
    const rgb = parseColor(color);
    const channels = [rgb.r, rgb.g, rgb.b].map((channel) => {
      const value = channel / 255;
      return value <= 0.03928
        ? value / 12.92
        : Math.pow((value + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
  };
  const contrastRatio = (a, b) => {
    const l1 = luminance(a);
    const l2 = luminance(b);
    const lighter = Math.max(l1, l2);
    const darker = Math.min(l1, l2);
    return (lighter + 0.05) / (darker + 0.05);
  };
  const readableTextColor = (backgroundColor) =>
    contrastRatio(backgroundColor, "#000000") >=
      contrastRatio(backgroundColor, "#ffffff")
      ? "#000000"
      : "#ffffff";
  const accentTextColor = (backgroundColor) => {
    if (contrastRatio(backgroundColor, "#ffffff") >= 3) {
      return "#ffffff";
    }
    if (contrastRatio(backgroundColor, "#000000") >= 3) {
      return "#000000";
    }
    return readableTextColor(backgroundColor);
  };
  const resolveColor = (settings) => {
    const color = settings?.Color || "green";
    let colors = colorMap[color];
    if (!colors && color === "diy" && settings.ColorDesc) {
      colors = settings.ColorDesc.split(",")
        .map((item) => item.trim())
        .filter(Boolean);
      if (colors.length === 1) colors = [colors[0], colors[0], colors[0]];
      if (colors.length === 2) colors = [colors[0], colors[1], colors[0]];
    }
    colors = colors || colorMap.green;
    return {
      primaryColorHover: colors[0],
      primaryColor: colors[1],
      primaryColorPressed: colors[2],
      primaryColorSuppl: colors[0],
    };
  };
  const createRole = (name, color) => ({
    [`--user-${name}-color`]: color,
    [`--user-${name}-bg-color`]: rgba(color, 0.14),
    [`--user-${name}-border-color`]: rgba(color, 0.34),
    [`--user-${name}-text-color`]: accentTextColor(color),
  });
  const buildTheme = (settings = {}) => {
    const isDark = settings.Theme === "dark";
    const transparent = clamp(Number(settings.Transparent) || 0, 0, 100);
    const surfaceAlpha = (100 - transparent) / 100;
    const readableAlpha = clamp(surfaceAlpha, isDark ? 0.72 : 0.78, 1);
    const baseBg = isDark ? "#000000" : "#ffffff";
    const colors = resolveColor(settings);
    const text1 = isDark ? "rgba(255, 255, 255, 0.88)" : "rgba(0, 0, 0, 0.88)";
    const text2 = isDark ? "rgba(255, 255, 255, 0.72)" : "rgba(0, 0, 0, 0.68)";
    const text3 = isDark ? "rgba(255, 255, 255, 0.48)" : "rgba(0, 0, 0, 0.46)";
    const primaryText = accentTextColor(colors.primaryColor);
    const vars = {
      "--user-primary-color": colors.primaryColor,
      "--user-primary-color-hover": colors.primaryColorHover,
      "--user-primary-color-pressed": colors.primaryColorPressed,
      "--user-primary-color-suppl": colors.primaryColorSuppl,
      "--user-primary-text-color": primaryText,
      "--user-primary-tcolor": primaryText,
      "--user-bg-color": `${baseBg}${alphaHex(surfaceAlpha)}`,
      "--user-bg-1-color": `${baseBg}${alphaHex(surfaceAlpha)}`,
      "--user-bg-2-color": rgba(baseBg, clamp(surfaceAlpha * 0.82, 0, 1)),
      "--user-bg-3-color": rgba(baseBg, clamp(surfaceAlpha * 0.64, 0, 1)),
      "--user-bg-filter-color": rgba(baseBg, isDark ? 0.2 : 0.26),
      "--user-glass-tint-color": isDark
        ? "rgba(0, 0, 0, 0.18)"
        : "rgba(255, 255, 255, 0.3)",
      "--user-glass-blur": `${transparent > 0 ? 3 : 0}px`,
      "--user-surface-color": rgba(baseBg, surfaceAlpha),
      "--user-surface-muted-color": rgba(baseBg, clamp(surfaceAlpha * 0.72, 0, 1)),
      "--user-readable-surface-color": rgba(baseBg, readableAlpha),
      "--user-control-color": isDark ? "rgba(255, 255, 255, 0.08)" : "rgba(0, 0, 0, 0.04)",
      "--user-control-hover-color": isDark ? "rgba(255, 255, 255, 0.13)" : "rgba(0, 0, 0, 0.07)",
      "--user-border-color": isDark ? "rgba(255, 255, 255, 0.18)" : "rgba(0, 0, 0, 0.13)",
      "--user-divider-color": isDark ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.08)",
      "--user-hover-color": isDark ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.06)",
      "--user-active-color": rgba(colors.primaryColor, isDark ? 0.26 : 0.16),
      "--user-text-color": text1,
      "--user-text-1-color": text1,
      "--user-text-2-color": text2,
      "--user-text-3-color": text3,
      "--user-text-muted-color": text3,
      "--user-round-enabled": settings.Round === "false" ? "0" : "1",
      ...createRole("success", "#18a058"),
      ...createRole("warning", "#f0a020"),
      ...createRole("error", "#d03050"),
      ...createRole("info", "#2080f0"),
    };
    return { vars, colors };
  };
  const applyThemeVars = (target, vars) => {
    const style = target?.style || target;
    if (!style || !vars) return;
    Object.entries(vars).forEach(([key, value]) => {
      if (value != null) style.setProperty(key, String(value));
    });
  };
  const theme = {
    current: null,
    build: buildTheme,
    apply(settings) {
      const data = settings || window.__personal || {};
      const current = {
        settings: data,
        ...buildTheme(data),
      };
      applyThemeVars(document.documentElement, current.vars);
      document.documentElement.dataset.theme =
        data.Theme === "dark" ? "dark" : "light";
      document.documentElement.dataset.lang = data.Lang || "zh";
      window.__personal = data;
      window.__personalTheme = current;
      theme.current = current;
      return current;
    },
    createColorRole(name, color) {
      const vars = createRole(name, color);
      applyThemeVars(document.documentElement, vars);
      return vars;
    },
  };
  window.addEventListener("message", (event) => {
    const data = event.data;
    if (data.action == "win-id-get" && data.data && data.data.id) {
      //获取当前窗口Id
      window.__winId = data.data.id;
      $v9os.webHost = data.data.webHost;
      $v9os.event.on("personalChange", function (settings) {
        const currentTheme = $v9os.theme.apply(settings);
        if (window.onPersonalChange) {
          window.onPersonalChange(settings, currentTheme);
        } else {
          window.__personal = settings;
        }
      });
    } else if (data.action == "iframe-invoke" && data.data) {
      //插件调用结果缓存
      if (data.data.taskId && invokeMap[data.data.taskId]) {
        invokeMap[data.data.taskId]({
          code: data.data.code,
          msg: data.data.msg,
          data: data.data.res,
        });
      }
    } else if (data.action == "iframe-event-on" && data.data) {
      //触发事件回调
      if (eventMap[data.data.eventName]) {
        eventMap[data.data.eventName].forEach((callback) => {
          callback(data.data.data);
        });
      }
    } else if (data.action == "context-menu-action" && data.data) {
      contextMenuOpened = false;
      const actionId = data.data.actionId;
      if (contextMenuActionMap[actionId]) {
        contextMenuActionMap[actionId].forEach((callback) => {
          callback(data.data);
        });
      }
      Object.keys(contextMenuActionMap).forEach((key) => {
        if (!key.endsWith("*") || !actionId.startsWith(key.slice(0, -1))) return;
        contextMenuActionMap[key].forEach((callback) => {
          callback(data.data);
        });
      });
    }
  });
  const tmpMsg = (color, text) => {
    const div = document.createElement("div");
    div.innerHTML = text;
    div.style =
      "background-color: #fff;position: fixed;top: 10px;left: calc(50vw - 150px);width: 300px;color: " +
      color +
      ";border: 1px solid #666;border-radius: 5px;padding: 5px;";
    document.body.appendChild(div);
    setTimeout(() => {
      div.remove();
    }, 3000);
  };
  const showSuccess = (msg) => {
    if (window.__winId) {
      $v9os.invoke("$msg.message", "success", msg);
    } else {
      if (window.$msg && $msg.message && $msg.message.success) {
        $msg.message.success(msg);
      } else {
        tmpMsg("green", msg);
      }
    }
  };
  const showError = (msg) => {
    if (window.__winId) {
      $v9os.invoke("$msg.message", "error", msg);
    } else {
      if (window.$msg && $msg.message && $msg.message.error) {
        $msg.message.error(msg);
      } else {
        tmpMsg("red", msg);
      }
    }
  };
  window.$v9os = {
    //前端反射跨域调用主程序前端功能
    invoke: (entity, method, ...param) => {
      return new Promise(function (success) {
        if (!window.__winId) {
          success({ code: -1, msg: "该方法必须在v9os iframe环境下使用" });
          return;
        }
        const taskId = `t${Date.now()}${Math.random().toString(36).slice(2)}`;
        invokeMap[taskId] = function (res) {
          delete invokeMap[taskId];
          success(res);
        };
        param = jsonClone(param);
        postToParent("iframe-invoke", {
          entity,
          method,
          param,
          winId: window.__winId,
          taskId,
        });
        setTimeout(() => {
          delete invokeMap[taskId];
          success({ code: -1, msg: "执行超时" });
        }, timeout * 1000);
      });
    },
    event: {
      on: function (eventName, callback) {
        if (!window.__winId) {
          return { code: -1, msg: "该方法必须在v9os iframe环境下使用" };
        }
        if (!eventMap[eventName]) {
          eventMap[eventName] = [];
        }
        if (typeof callback === "function" && !eventMap[eventName].includes(callback)) {
          eventMap[eventName].push(callback);
        }
        if (!hostEventSubscribed[eventName]) {
          hostEventSubscribed[eventName] = true;
          postToParent("iframe-event-on", { eventName, winId: window.__winId });
        }
        return { code: 0, msg: "success" };
      },
      off: function (eventName, callback) {
        if (eventMap[eventName]) {
          if (callback) {
            eventMap[eventName] = eventMap[eventName].filter(
              (x) => x != callback,
            );
            if (eventMap[eventName].length == 0) {
              delete eventMap[eventName];
            }
          } else {
            delete eventMap[eventName];
          }
        }
        if (!eventMap[eventName] && hostEventSubscribed[eventName]) {
          delete hostEventSubscribed[eventName];
          if (window.__winId) {
            postToParent("iframe-event-off", { eventName, winId: window.__winId });
          }
        }
        return { code: 0, msg: "success" };
      },
      emit: function (eventName, data) {
        postToParent(eventName, jsonClone(data));
        return { code: 0, msg: "success" };
      },
    },
    contextMenu: {
      show: async function (options) {
        if (!window.__winId) {
          return { code: -1, msg: "该方法必须在v9os iframe环境下使用" };
        }
        const res = await $v9os.invoke("$contextMenu", "showFromPlugin", window.__winId, options);
        contextMenuOpened = res && res.code === 0;
        return res;
      },
      close: async function () {
        if (!contextMenuOpened) {
          return { code: 0, msg: "success" };
        }
        contextMenuOpened = false;
        return await $v9os.invoke("$contextMenu", "close");
      },
      onAction: function (actionName, callback) {
        if (!contextMenuActionMap[actionName]) {
          contextMenuActionMap[actionName] = [callback];
        } else {
          contextMenuActionMap[actionName].push(callback);
        }
        return { code: 0, msg: "success" };
      },
      offAction: function (actionName, callback) {
        if (contextMenuActionMap[actionName]) {
          if (callback) {
            contextMenuActionMap[actionName] = contextMenuActionMap[actionName].filter((x) => x != callback);
            if (contextMenuActionMap[actionName].length == 0) {
              delete contextMenuActionMap[actionName];
            }
          } else {
            delete contextMenuActionMap[actionName];
          }
        }
        return { code: 0, msg: "success" };
      },
      setClipboard: async function (data) {
        return await $v9os.invoke("$contextMenu", "setClipboard", data);
      },
      getClipboard: async function () {
        return await $v9os.invoke("$contextMenu", "getClipboard");
      },
    },
    theme,
    auth: {
      has: function (path) {
        const auths = $v9os.__userAuths || [];
        $v9os.invoke("$user.auths", "value").then((v) => {
          const list = v && v.code === 0 && v.data && Array.isArray(v.data.auths) ? v.data.auths : [];
          $v9os.__userAuths = list;
        })
        return auths.includes("all") || auths.includes(path);
      },
    },
    api: {
      webDataPost: async function (module, action, param, showType = "err") {
        return await $v9os.api.pluginPost("../api/webdata", module + "/" + action, param, showType);
      },
      pluginPost: async function (module, action, param, showType = "err", fn) {
        let res;
        let tmp = await $v9os.invoke(
          "window",
          "_pluginPostData",
          module,
          action,
          param,
        );
        if (tmp.code == 0) {
          res = tmp.data;
        } else {
          res = { code: tmp.code, msg: tmp.msg };
        }
        if (fn) {
          fn(res.msg);
        }
        if (showType == "json") {
          return res;
        }
        if (res && res.code === 0) {
          if (showType === "ok" || showType === "okerr") {
            showSuccess(res.msg);
          }
          if (res.data) {
            return res.data;
          }
          return true;
        }
        if (showType === "err" || showType === "okerr") {
          showError(res.msg);
        }
        return false;
      },
    },
    host,
    msg: {
      alert: async function (content, title) {
        if (!title) {
          title = "提示"
        }
        let res = await $v9os.invoke("$msg.util", "alert", content, title);
        return res && res.code == 0 && res.data;
      },
      confirm: async function (content, title) {
        if (!title) {
          title = "提示"
        }
        let res = await $v9os.invoke("$msg.util", "confirm", content, title);
        return res && res.code == 0 && res.data;
      },
      prompt: async function (content, title) {
        if (!title) {
          title = "提示"
        }
        let res = await $v9os.invoke("$msg.util", "prompt", content, title);
        return res && res.code == 0 && res.data && res.data.trim();
      },
      success: showSuccess,
      error: showError
    },
    file: {
      common: function (relative, title, expand, ext, name, save) {
        return new Promise(function (success) {
          const onExpandChange = (msg) => {
            if (msg?.type !== "fileSelected") {
              return;
            }
            if (msg.winId) {
              $v9os.invoke(
                "$wins",
                "closeWindow",
                msg.winId
              );
            }
            $v9os.event.off("expand-change", onExpandChange);
            if (msg.from == "win-close" || !msg.confirm) {
              success(false);
              return;
            }
            success(msg.data);
          }
          $v9os.event.on("expand-change", onExpandChange);
          if (relative != "true") {
            relative = "false";
          }
          if (!ext) {
            ext = ""
          }
          if (!name) {
            name = ""
          }
          $v9os.invoke(
            "$wins",
            "addWindow",
            {
              width: 900,
              height: 620,
              title: title ? title : "请选择",
              iframeUrl: `${$v9os.host}/page/file_system/?page=explorer&expand=${expand}&relative=${relative}&ext=${encodeURIComponent(ext)}&name=${encodeURIComponent(name)}&save=${save}`,
            },
            window.__winId,
            {
              name: "expand-change",
              data: {
                type: "fileSelected"
              }
            }
          );
        });
      },
      selectLongDir: function (relative, title) {
        return this.common(relative, title ? title : "请选择目录", "selectDir")
      },
      selectFile: function (relative, title, ext, save) {
        //ext = png,jpg
        return this.common(relative, title ? title : "请选择文件", "selectFile", ext, null, save == "true" ? "true" : "false")
      },
      selectLongFile: function (relative, title, ext) {
        //ext = png,jpg
        return this.common(relative, title ? title : "请选择文件", "selectFile", ext, null, "")
      },
      saveFile: async function (title, name, blob) {
        const postData = await this.common(null, title ? title : "请选择保存目录", "saveFile", null, name);
        if (!postData) {
          return false;
        }
        return this.saveFileNoSelect(postData, blob)
      },
      saveFileNoSelect: async function (postData, blob) {
        const resp = await fetch(postData.url, {
          "headers": postData.headers,
          "body": blob,
          "method": postData.method
        });
        let err = "保存失败"
        if (resp.ok) {
          const json = await resp.json()
          if(json.code == 0){
            $v9os.msg.success("保存成功");
            return true;
          }
          err = json.msg || err
        }
        $v9os.msg.error(err);
        return false;
      }
    }
  };
  const closeContextMenuFromFrame = function () {
    if (window.$v9os && $v9os.contextMenu) {
      $v9os.contextMenu.close();
    }
  };
  document.addEventListener("mousedown", closeContextMenuFromFrame);
  document.addEventListener("touchstart", closeContextMenuFromFrame);
  if (window.__personal) {
    theme.apply(window.__personal);
  }
  postToParent("win-id-get");
})();
