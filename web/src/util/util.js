import emitter from "@/util/event.js";
import encodeBase64 from "@/util/base64.js";
export function uuid() {
  return "xxxxxyxxxxyxxxxxyxxxxxxxxyxxxxxx".replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
let apiHost = null;
export async function getApiHost() {
  if (apiHost && apiHost.host) {
    return apiHost.host;
  }
  await postData("health", "", null, "");
  return apiHost.host;
}
export async function postData(module, action, param, showType = "err", fn) {
  //err  okerr  ok
  const resp = await postReq(module, action, param);
  if (showType == "resp") {
    return resp;
  }
  const res = await resp.json().catch((msg) => {
    return {
      code: -1,
      msg: msg.toString(),
    };
  });
  if (fn) {
    fn(res.msg);
  }
  if (showType == "json") {
    return res;
  }
  if (res && res.code === 0) {
    if (showType === "ok" || showType === "okerr") {
      $msg.message.success(res.msg);
    }
    if (res.data) {
      return res.data;
    }
    return true;
  }
  if (showType === "err" || showType === "okerr") {
    $msg.message.error(res.msg);
  }
  return false;
}

export async function postStreamData(module, action, param, callback, showType = "err") {
  const resp = await postReq(module, action, param);
  if (!resp.ok || !resp.body) {
    const res = await resp.json().catch((msg) => ({
      code: -1,
      msg: msg.toString(),
    }));
    if (showType === "err" || showType === "okerr") {
      $msg.message.error(res.msg);
    }
    return false;
  }
  const reader = resp.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let hasError = false;
  const consume = (text) => {
    buffer += text.replace(/\r\n/g, "\n");
    const frames = buffer.split("\n\n");
    buffer = frames.pop() || "";
    for (const frame of frames) {
      const dataText = frame
        .split("\n")
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.replace(/^data:\s?/, ""))
        .join("\n");
      if (!dataText) {
        continue;
      }
      let data = null;
      try {
        data = JSON.parse(dataText);
      } catch (err) {
        data = { code: -1, msg: err.message, raw: dataText };
      }
      if (data?.code === -1) {
        hasError = true;
      }
      if (callback) {
        callback(data);
      }
    }
  };
  while (true) {
    const { value, done } = await reader.read();
    if (done) {
      break;
    }
    consume(decoder.decode(value, { stream: true }));
  }
  const tail = decoder.decode();
  if (tail) {
    consume(tail);
  }
  if (buffer.trim()) {
    consume("\n\n");
  }
  return !hasError;
}

const refreshToken = { need: 0, fn: null, token: "" };
async function postReq(module, action, param) {
  //err  okerr  ok
  if (!apiHost) {
    apiHost = await fetch(window.__APP_WEB_BASE_PATH + "/service.json")
      .then((res) => res.json())
      .catch((err) => ({ host: window.location.origin }));
    if (!apiHost) {
      apiHost = { host: window.location.origin };
    }
    if (!apiHost.host) {
      apiHost.host = window.location.origin;
    }
    delete window.__APP_WEB_BASE_PATH;
  }
  const headers = {
    lang: localStorage.getItem("lang"),
  };
  if (localStorage.getItem("token")) {
    headers["Authorization"] = "Bearer " + localStorage.getItem("token");
  }
  const method = param ? "POST" : "GET";
  let body = null;
  if (param) {
    if (param instanceof FormData) {
      body = param;
    } else {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(param);
    }
  }
  let url = `${apiHost.host}/api/${module}/${action}`;
  if (module == "health") {
    url = `${apiHost.host}/${module}`;
  }
  const res = await fetch(url, {
    method,
    body,
    headers,
  });
  if (window["$user"] && refreshToken.need == 0) {
    const token = localStorage.getItem("refreshToken");
    if (res.headers.get("TokenLast") == "true" && token) {
      refreshToken.need = 1;
      refreshToken.token = token;
    } else if (res.status == 401 && token) {
      refreshToken.need = 2;
      refreshToken.token = token;
    }
    if (refreshToken.need > 0) {
      refreshToken.fn = async function () {
        const resp = await postData("user", "token", {
          token: refreshToken.token
        }, "resp");
        if (resp.status == 200) {
          const jsonRes = await resp.json();
          let tokenRes = jsonRes.code == 0 ? jsonRes.data : null;
          if (tokenRes) {
            await $user.setToken(tokenRes);
            $user.loadUser();
          }
          if (refreshToken.need == 2 && !tokenRes) {
            localStorage.removeItem("refreshToken");
            localStorage.removeItem("token");
            emitter.emit("token-expired");
          }
          refreshToken.need = 0;
          setTimeout(() => {
            refreshToken.need = 0;
          }, 300000);
        } else {
          refreshToken.need = 0;
        }
      };
      const timeMs = Math.floor(Math.random() * 1001);
      setTimeout(() => {
        const fn = refreshToken.fn;
        delete refreshToken.fn;
        if (fn) {
          fn();
        }
      }, timeMs);
    }
  }
  return res;
}

export async function postBlob(module, action, param, showType = "err") {
  //err  okerr  ok
  const resp = await postReq(module, action, param);
  if (resp.status != 200 && resp.status != 206) {
    const res = await resp.json().catch((msg) => {
      return {
        code: -1,
        msg: msg.toString(),
      };
    });
    if (showType === "err" || showType === "okerr") {
      $msg.message.error(res.msg);
    }
    return null;
  }
  const name = decodeURIComponent(resp.headers.get("File-Name"));
  let errMsg = $t("common.all.fileNotFound");
  const blob = await resp.blob().catch((msg) => {
    errMsg = msg;
    return null;
  });
  if (blob == null && (showType === "err" || showType === "okerr")) {
    $msg.message.error(errMsg);
    return null;
  }
  if (showType === "ok" || showType === "okerr") {
    $msg.message.success($t("common.all.fileDownloadSuccess"));
  }
  return { name, blob };
}

export function getWinSize(pHWidth) {
  if (!pHWidth) {
    pHWidth = 1500;
  }
  const maxWidth = Math.min(window.innerWidth, pHWidth);
  const maxHeight = Math.min(window.innerHeight, pHWidth * 5 / 8);
  const ratio = 8 / 5;
  const bl = 0.8;
  const byWidth = {
    width: maxWidth * bl,
    height: (maxWidth * bl) / ratio,
  };
  const byHeight = {
    width: maxHeight * bl * ratio,
    height: maxHeight * bl,
  };
  return byWidth.width > byHeight.width ? byHeight : byWidth;
}
export function elementInMe(me, ele) {
  let flag = false;
  while (true) {
    if (!ele || ele == document.body) {
      flag = false;
      break;
    }
    if (ele == me) {
      flag = true;
      break;
    } else {
      ele = ele.parentElement;
    }
  }
  return flag;
}
const that = {};
export function delayAction(tjFn, acFn, maxDelay) {
  if (!maxDelay) {
    maxDelay = 60 * 60 * 1000;
  }
  if (maxDelay < 10000) {
    maxDelay = 10000;
  }
  let key = "da" + uuid();
  let timeKey = "time" + key;
  that[timeKey] = Date.now();
  that[key] = function () {
    if (Date.now() - that[timeKey] > maxDelay) {
      delete that[key];
      delete that[timeKey];
    } else {
      if (tjFn()) {
        delete that[key];
        delete that[timeKey];
        acFn();
      } else {
        setTimeout(that[key], 100);
      }
    }
  };
  that[key]();
}
export function emojiToBase64(emoji) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
    <text  x="50"  y="50" font-size="80" text-anchor="middle" dominant-baseline="central" font-family="'Segoe UI Emoji', 'Apple Color Emoji', sans-serif" fill="black">${emoji}</text></svg>`;
  const base64 = 'data:image/svg+xml;base64,' + encodeBase64(svg);
  return base64;
}
export function getComplementaryColor(backgroundColor) {
  const tempEl = document.createElement('div');
  tempEl.style.color = backgroundColor;
  document.body.appendChild(tempEl);
  const computedColor = getComputedStyle(tempEl).color;
  document.body.removeChild(tempEl);

  const rgbMatch = computedColor.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/);
  if (!rgbMatch) return '#000000';

  const bgColor = {
    r: parseInt(rgbMatch[1]),
    g: parseInt(rgbMatch[2]),
    b: parseInt(rgbMatch[3])
  };
  const getLuminance = (rgb) => {
    const sRGB = [rgb.r, rgb.g, rgb.b].map(c => {
      c = c / 255;
      return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * sRGB[0] + 0.7152 * sRGB[1] + 0.0722 * sRGB[2];
  };
  const getContrastRatio = (color1, color2) => {
    const lum1 = getLuminance(color1);
    const lum2 = getLuminance(color2);
    const brightest = Math.max(lum1, lum2);
    const darkest = Math.min(lum1, lum2);
    return (brightest + 0.05) / (darkest + 0.05);
  };
  const black = { r: 0, g: 0, b: 0 };
  const white = { r: 255, g: 255, b: 255 };
  const contrastWithBlack = getContrastRatio(bgColor, black);
  const contrastWithWhite = getContrastRatio(bgColor, white);
  const minContrastRatio = 4.5;
  if (contrastWithBlack >= minContrastRatio && contrastWithWhite >= minContrastRatio) {
    return contrastWithBlack > contrastWithWhite ? '#000000' : '#ffffff';
  }
  if (contrastWithBlack >= minContrastRatio) return '#000000';
  if (contrastWithWhite >= minContrastRatio) return '#ffffff';
  const luminance = (0.299 * bgColor.r + 0.587 * bgColor.g + 0.114 * bgColor.b) / 255;
  return luminance > 0.5 ? '#000000' : '#ffffff';
}
const fnMap = {}
export async function lockFn(key, fn) {
  if (fnMap[key]) {
    return { flag: false, msg: "please try again" };
  }
  fnMap[key] = 1;
  try {
    return { flag: true, data: await fn() };
  } catch (e) {
    return { flag: false, msg: e.message }
  } finally {
    delete fnMap[key];
  }
}
