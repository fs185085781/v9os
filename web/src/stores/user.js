import { reactive, ref, defineAsyncComponent, computed } from "vue";
import { defineStore } from "pinia";
import { darkTheme, zhCN, dateZhCN, enUS, dateEnUS } from "naive-ui";
import { loadLocaleMessages } from "@/locales";
import { postData, lockFn, postBlob, absoluteUrl } from "@/util/util";
import { MessageOutlined } from "@vicons/antd";
import { DataUsage20Regular } from "@vicons/fluent";
import { renderIcon } from "@/util/icon";
import { getBigData, setBigData } from "@/util/bigdata";
import emitter from "@/util/event.js";
import {
  applyThemeVars,
  buildPersonalTheme,
} from "@/util/personalTheme.js";
const langPatterns = {
  zh: {
    lang: zhCN,
    date: dateZhCN,
  },
  en: {
    lang: enUS,
    date: dateEnUS,
  },
};
const registeredBuiltinApps = reactive([]);
export const useStore = defineStore("user", () => {
  const user = reactive({ ID: 0 });
  const system = reactive({
    Shutdown: false,
    Open: false,
    Wakeup: false,
  });
  const webSettings = ref({});
  const auths = ref({ auths: [], init: false });
  const settings = reactive({
    Mode: null,
    Theme: null,
    Lang: null,
    Lang1: null,
    Lang2: null,
    Round: "true",
    Color: "green",
    Font: "default",
    ThemeOverride: null,
    ThemeVars: null,
    Mourning: "false",
    DockApps: '[]',
    SoundVolume: Number(localStorage.getItem("v9os.soundVolume") || 60),
    Transparent: 0,
  });
  //Color,Round,Lang,Mode,Theme,Font,Transparent
  const loadUser = async () => {
    let res = await postData("system", "settingsGet", {}, "");
    if (res) {
      webSettings.value = res;
      if (window.__MAIN_FEST_SAVE) {
        window.__MAIN_FEST_SAVE({ logo: absoluteUrl(res.Logo), name: res.Title });
        delete window.__MAIN_FEST_SAVE;
      }
      document.title =
        webSettings.value.Title + "-" + webSettings.value.Subtitle;
    }
    let hasLogin = false;
    let userInfo = await postData("user", "info", {}, "");
    if (userInfo) {
      hasLogin = true;
      for (const key in userInfo) {
        user[key] = userInfo[key];
      }
    }
    if (hasLogin) {
      res = await postData("user", "auths", {}, "");
      if (res) {
        auths.value = { auths: res ? res : [], init: true };
      }
    }
    res = await postData("user", "settings", {}, "");
    if (res) {
      const lastMode = settings.Mode;
      const fields = [
        "Color",
        "ColorDesc",
        "Round",
        "Lang",
        "Mode",
        "Theme",
        "Font",
        "DefaultWallpaper",
        "DefaultWallpaperType",
        "DockApps",
        "Transparent",
      ];
      fields.forEach((field) => {
        if (res[field] != null) settings[field] = res[field];
      });
      setTheme(settings.Theme);
      setLang(settings.Lang);
      setUiMode(settings.Mode, false, lastMode);
      setRound(settings.Round);
      setColor(settings.Color);
      setFonts(settings.Font);
      setMourning(webSettings.value.Mourning);
    }
    const isV9osApp = window.$util && $util.proxy && $util.proxy.setToken;
    if (isV9osApp) {
      const res = await postData("user", "proxyToken", { host: absoluteUrl() }, "");
      if (res) {
        $util.proxy.setToken(res.proxy_host, res.proxy_token);
      }
    }
  };
  loadUser();
  const saveUserSettings = async () => {
    const data = {};
    const fields = [
      "Color",
      "ColorDesc",
      "Round",
      "Lang",
      "Mode",
      "Theme",
      "Font",
      "DefaultWallpaper",
      "DefaultWallpaperType",
      "DockApps",
      "Transparent",
    ];
    fields.forEach((field) => {
      data[field] = settings[field];
    });
    if (await postData("user", "saveSettings", data, "okerr")) {
      emitter.emit("personalChange", getPersonalPayload());
    }
  };
  const applyPersonalTheme = () => {
    const theme = buildPersonalTheme(settings);
    applyThemeVars(document.documentElement, theme.vars);
    settings.ThemeVars = theme.vars;
    settings.ThemeOverride = theme.naiveThemeOverrides;
    return theme;
  };
  const getPersonalPayload = () => {
    const theme = applyPersonalTheme();
    return {
      Theme: settings.Theme,
      Color: settings.Color,
      Round: settings.Round,
      Lang: settings.Lang,
      Mode: settings.Mode,
      Font: settings.Font,
      ColorDesc: settings.ColorDesc,
      Transparent: settings.Transparent,
      ThemeVars: theme.vars,
      ThemeOverride: theme.naiveThemeOverrides,
    };
  };
  const setTheme = (theme, save) => {
    if (settings.Mourning == "true") {
      return;
    }
    settings.Theme = theme;
    if (theme == "dark") {
      settings.NaiveTheme = darkTheme;
    } else {
      settings.NaiveTheme = null;
    }
    applyPersonalTheme();
    if (save) {
      saveUserSettings();
    }
  };
  const setTransparent = (transparent, save) => {
    settings.Transparent = Math.max(0, Math.min(100, Number(transparent) || 0));
    applyPersonalTheme();
    if (save) {
      saveUserSettings();
    }
  };
  const setWallpaper = (wallpaper, type, save) => {
    settings.DefaultWallpaper = wallpaper;
    settings.DefaultWallpaperType = type;
    if (save) {
      saveUserSettings();
    }
  };
  const setLang = async (lang, save) => {
    localStorage.setItem("lang", lang);
    const tmp = langPatterns[lang];
    if (tmp) {
      settings.Lang1 = tmp.lang;
      settings.Lang2 = tmp.date;
      await loadLocaleMessages(lang);
      emitter.emit("lang-change", lang);
    }
    if (save) {
      settings.Lang = lang;
      saveUserSettings();
    }
  };
  const setUiMode = (mode, save, lastMode) => {
    if (!["macos", "backend", "win10", "deepin"].includes(mode)) {
      $msg.message.error("当前模式暂未支持");
      return;
    }
    if (save) {
      if (!['win10', 'macos', 'deepin'].includes(mode)) {
        settings.Transparent = 0;
      }
      settings.Mode = mode;
      saveUserSettings();
    }
    if (mode != lastMode) {
      if ($wins.windows.length > 0) {
        $wins.windows.splice(0, $wins.windows.length);
      }
    }
    setFonts(settings.Font);
  };
  const setRound = (round, save) => {
    settings.Round = round;
    applyPersonalTheme();
    if (save) {
      saveUserSettings();
    }
  };
  const setColor = (color, save) => {
    settings.Color = color;
    applyPersonalTheme();
    if (save) {
      saveUserSettings();
    }
  };
  const setMourning = (mourning) => {
    settings.Mourning = mourning;
    if (settings.Mourning == "true") {
      settings.TmpTheme = settings.Theme;
      settings.Theme = null;
      document.querySelector("html").className = "mourning";
    } else {
      setTheme(settings.TmpTheme ? settings.TmpTheme : settings.Theme);
      document.querySelector("html").className = "";
    }
  };
  const setSoundVolume = (volume) => {
    const value = Math.max(0, Math.min(100, Number(volume) || 0));
    settings.SoundVolume = value;
    localStorage.setItem("v9os.soundVolume", String(value));
    document.querySelectorAll("audio, video").forEach((media) => {
      media.volume = value / 100;
    });
  };
  const setFonts = async (fonts, save) => {
    fonts = String(fonts || "default");
    if (save) {
      settings.Font = fonts;
      saveUserSettings();
    }
    let ele = document.querySelector("#v9os-fonts");
    if (!ele) {
      ele = document.createElement("style");
      ele.id = "v9os-fonts";
      document.head.appendChild(ele);
    }
    const tmp = await lockFn("fonts-get", async () => {
      const key = `fonts-${fonts}${fonts == "default" ? "-" + settings.Mode : ""}`;
      let fontsData = await getBigData(key);
      if (fontsData === undefined || (fontsData && fontsData.time && Date.now() - fontsData.time > 5 * 60 * 1000)) {
        fontsData = {}
        let res = await postBlob("appstore", "read_fonts", { ui: settings.Mode, font: fonts }, "")
        if (res && res.blob) {
          fontsData.blob = res.blob
        } else {
          fontsData.time = Date.now()
        }
        await setBigData(key, fontsData);
      }
      if (fontsData.blob) {
        return URL.createObjectURL(fontsData.blob);
      }
      return null;
    })
    if (tmp.flag && tmp.data) {
      const url = tmp.data;
      ele.innerHTML = `@font-face {
        font-family: 'v9os';
        src: url('${url}') format('woff2'),url('${url}') format('woff'),url('${url}') format('truetype');
        font-weight: normal;
        font-style: normal;
      }
      body{
        font-family: 'v9os', sans-serif !important;
      }  
    `;
    } else {
      ele.innerHTML = "";
    }
  };
  const getDockApps = () => {
    try {
      return JSON.parse(settings.DockApps || "[]");
    } catch {
      return [];
    }
  };
  const addDockApp = (code) => {
    const apps = getDockApps();
    if (!apps.includes(code)) {
      apps.push(code);
      settings.DockApps = JSON.stringify(apps);
      saveUserSettings();
      emitter.emit("dock-apps-change");
    }
  };
  const removeDockApp = (code) => {
    const apps = getDockApps();
    const idx = apps.indexOf(code);
    if (idx !== -1) {
      apps.splice(idx, 1);
      settings.DockApps = JSON.stringify(apps);
      saveUserSettings();
      emitter.emit("dock-apps-change");
    }
  };
  const setToken = async (token) => {
    const sz = token.split("|")
    localStorage.setItem("token", sz[0]);
    if (sz[1]) {
      localStorage.setItem("refreshToken", sz[1]);
    }
  };
  const logout = () => {
    window.$websocket?.removeAllClients?.();
    localStorage.removeItem("token");
    localStorage.removeItem("refreshToken");
    emitter.emit("token-expired");
  };
  const getDesktopApps = async () => {
    const rows = await postData("user", "desktop_apps", {}, "");
    if (!Array.isArray(rows)) return [];
    const myApps = await getMyApps();
    const appMap = {};
    myApps.forEach((app) => {
      appMap[app.code] = app;
    });
    return rows.map((row) => {
      const app = appMap[row.Code] || {};
      const appType = row.AppType || "iframe";
      return {
        id: row.ID,
        ID: row.ID,
        code: row.Code || `desktop_${row.ID}`,
        name: row.Title || app.name || app.code,
        icon: row.Icon || app.icon || defaultIcon("appstore"),
        type: appType === "iframe" ? "url" : appType,
        windowId: app.windowId || "",
        url: row.Url || app.url || "",
        AppType: row.AppType,
        Code: row.Code,
        Title: row.Title,
        Icon: appType === "system" ? "" : row.Icon,
        Url: appType === "system" ? "" : row.Url,
        Sort: row.Sort,
      };
    });
  };
  const defaultIcon = function (icon) {
    return `/assets/${["macos", "win10", "deepin"].includes(settings.Mode) ? settings.Mode : "common"}/img/icons/${icon}.png`;
  }
  const saveDesktopApp = async (app) => {
    const res = await postData("user", "saveDesktopApp", app, "okerr");
    if (res) emitter.emit("desktop-apps-change", {});
    return res;
  };
  const deleteDesktopApp = async (id) => {
    const res = await postData("user", "deleteDesktopApp", { ID: id }, "okerr");
    if (res) emitter.emit("desktop-apps-change", {});
    return res;
  };
  const addDesktopApp = async (app) => {
    const appType = app.type === "system" ? "system" : app.type === "plugin" ? "plugin" : "iframe";
    return saveDesktopApp({
      Icon: appType === "system" ? "" : (app.icon || app.Icon || defaultIcon("appstore")),
      Title: app.name?.value || app.name || app.title || app.Title || app.code || "快捷方式",
      AppType: appType,
      Code: app.code || app.Code || "",
      Url: appType === "system" ? "" : (app.url || app.Url || ""),
    });
  };
  const registerBuiltinApps = (apps) => {
    if (!Array.isArray(apps)) {
      return;
    }
    apps.forEach((app) => {
      if (!app?.code) {
        return;
      }
      const index = registeredBuiltinApps.findIndex((item) => item.code == app.code);
      if (index >= 0) {
        registeredBuiltinApps.splice(index, 1, app);
      } else {
        registeredBuiltinApps.push(app);
      }
    });
  };
  let moduleList = null;
  const getPluginApps = async () => {
    const apps = [];
    if (!moduleList) {
      moduleList = await postData("user", "auth-modules", {}, "");
    }
    for (const mod of moduleList) {
      let path = `/page/${mod.code}/`;
      if (mod.type === 2) {
        path = `/api/webplugin/${mod.code}/`;
      } else if (mod.type === 3) {
        path = `/api/thirdplugin/${mod.code}/`;
      } else if (mod.type === 4) {
        path = mod.accessUrl || "";
      }
      mod.icon = mod.icon ? absoluteUrl(mod.icon) : defaultIcon("appstore");
      apps.push({
        code: mod.code,
        name: mod.name,
        icon: mod.icon || defaultIcon("appstore"),
        type: "plugin",
        url: absoluteUrl(path),
        editExts: mod.editExts,
        openExts: mod.openExts,
        expandExts: mod.expandExts,
      });
    }
    return apps;
  }

  const getMyApps = async () => {
    const builtinApps = {
      __kernel__: {
        code: "__kernel__",
        name: $tc("admin.title"),
        icon: renderIcon(DataUsage20Regular, 40),
        type: "system",
        url: defineAsyncComponent(
          () => import("@/components/common/views/admin.vue"),
        ),
      },
      __appstore__: {
        code: "__appstore__",
        name: computed(() => $t("common.appstore.title")),
        icon: defaultIcon("appstore"),
        type: "system",
        url: defineAsyncComponent(() => import("@/components/common/modules/appstore/AppStore.vue")),
      },
      __chat__: {
        code: "__chat__",
        name: "聊天与通知",
        icon: renderIcon(MessageOutlined, 40),
        type: "system",
        windowId: "v9os-chat-win",
        url: defineAsyncComponent(() => import("@/components/common/component/user/chat/chat.vue")),
      }
    };
    registeredBuiltinApps.forEach((app) => {
      builtinApps[app.code] = app;
    });
    const myauths = auths.value?.auths || [];
    const apps = [];
    if (builtinApps.__settings__) {
      apps.push(builtinApps.__settings__);
    }
    if (!myauths.length) { return apps; }
    const isAll = myauths.includes("all");
    if (isAll || myauths.some((a) => a.startsWith("/api/"))) {
      apps.push(builtinApps.__kernel__);
    }
    if (isAll || myauths.some((a) => a.startsWith("/api/appstore/"))) {
      apps.push(builtinApps.__appstore__);
    }
    if (isAll || myauths.some((a) => a.startsWith("/api/chat_ee/") || a.startsWith("/api/chat/"))) {
      apps.push(builtinApps.__chat__);
    }
    const pluginApps = await getPluginApps();
    for (const mod of pluginApps) {
      apps.push({
        code: mod.code,
        name: mod.name,
        icon: mod.icon,
        type: mod.type,
        url: mod.url
      });
    }
    return apps;

  }
  window.$user = {
    user,
    system,
    webSettings,
    settings,
    auths,
    setTheme,
    setLang,
    setUiMode,
    setRound,
    setColor,
    setFonts,
    setWallpaper,
    setTransparent,
    loadUser,
    setMourning,
    setSoundVolume,
    setToken,
    logout,
    getDockApps,
    registerBuiltinApps,
    getMyApps,
    getPluginApps,
    getDesktopApps,
    saveDesktopApp,
    deleteDesktopApp,
    addDesktopApp,
    addDockApp,
    removeDockApp,
    defaultIcon
  };
  return window.$user;
});
