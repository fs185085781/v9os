import { defineAsyncComponent } from "vue";
import { LinkOutlined } from "@vicons/antd";
import { renderIcon } from "@/util/icon";

const DesktopShortcutWin = defineAsyncComponent(() => import("./DesktopShortcutWin.vue"));

export function desktopShortcutDataFromApp(app = {}) {
  const appType = app.AppType || (app.type === "url" ? "iframe" : app.type) || "iframe";
  const normalizedType = appType === "system" || appType === "plugin" ? appType : "iframe";
  return {
    ID: app.ID || app.id || 0,
    Title: app.Title || app.name?.value || app.name || app.title || app.code || "快捷方式",
    Icon: normalizedType === "system" ? "" : (app.Icon || app.icon || ""),
    AppType: normalizedType,
    Code: app.Code || app.code || "",
    Url: normalizedType === "system" ? "" : (app.Url || app.url || ""),
    Sort: Number(app.Sort || app.sort || 0),
  };
}

export function openDesktopShortcutWin(app = {}, options = {}) {
  $wins.addWindow({
    width: 520,
    height: 360,
    title: options.title || "桌面快捷方式",
    icon: renderIcon(LinkOutlined, 40),
    component: DesktopShortcutWin,
    data: desktopShortcutDataFromApp(app),
  });
}
