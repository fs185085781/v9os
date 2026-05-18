import { postData } from "@/util/util";

const providerMessageType = {
  qq: "qq-code",
  weixin: "wx-code",
};

const providerName = {
  qq: "QQ",
  weixin: "微信",
};

export async function getOAuthProviderConfig(provider) {
  const configs = await postData("user_ee", "oauth_config", {}, "err");
  const cfg = configs?.[provider];
  const name = providerName[provider] || "第三方";
  if (!cfg?.enabled) {
    $msg.message.warning(`${name}登录未启用`);
    return null;
  }
  if (!cfg.appId) {
    $msg.message.warning(`${name} AppId未配置`);
    return null;
  }
  return cfg;
}

export function buildOAuthRedirectUri(provider, scene) {
  const url = new URL(window.location.href);
  url.searchParams.set("oauth_provider", provider);
  url.searchParams.set("oauth_scene", scene || "login");
  return url.toString();
}

export async function openOAuthWindow(provider, scene = "login") {
  const cfg = await getOAuthProviderConfig(provider);
  if (!cfg) {
    return false;
  }
  const buildUrl = window.__V9OS_OAUTH_BUILD_URL;
  if (typeof buildUrl !== "function") {
    $msg.message.warning("第三方授权地址生成器未配置");
    return false;
  }
  const redirectUri = buildOAuthRedirectUri(provider, scene);
  const authorizeUrl = buildUrl(provider, cfg, redirectUri, scene);
  if (!authorizeUrl) {
    $msg.message.warning("第三方授权地址未配置");
    return false;
  }
  window.open(authorizeUrl, "_blank", "width=600,height=560");
  return true;
}

export function readOAuthCallbackFromLocation() {
  const url = new URL(window.location.href);
  const provider = url.searchParams.get("oauth_provider") || url.searchParams.get("provider");
  const code = url.searchParams.get("code");
  const scene = url.searchParams.get("oauth_scene") || "login";
  if (!provider || !code || !providerMessageType[provider]) {
    return null;
  }
  return { provider, code, scene, type: providerMessageType[provider] };
}

export function notifyOAuthCallbackIfNeeded() {
  const data = readOAuthCallbackFromLocation();
  if (!data) {
    return null;
  }
  if (window.opener && !window.opener.closed) {
    data.notified = true;
    window.opener.postMessage(data, window.location.origin);
    window.close();
  } else if (window.parent && window.parent !== window) {
    data.notified = true;
    window.parent.postMessage(data, window.location.origin);
  }
  return data;
}

export function listenOAuthCode(provider, callback) {
  const messageType = providerMessageType[provider];
  const handler = (event) => {
    if (event.origin !== window.location.origin) {
      return;
    }
    const data = event.data || {};
    if (data.provider === provider && data.code) {
      callback(data.code, data);
      return;
    }
    if (data.type === messageType && data.code) {
      callback(data.code, data);
    }
  };
  window.addEventListener("message", handler);
  return () => window.removeEventListener("message", handler);
}
