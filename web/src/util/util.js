import emitter from "@/util/event.js";
export function uuid() {
  return "xxxxxyxxxxyxxxxxyxxxxxxxxyxxxxxx".replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
let apiHost = null;
//本项目首次执行postData在语言加载,因此absoluteUrl方法在全局范围内(语言于main.js加载)都是安全的
export function absoluteUrl(url) {
  if (!url) {
    url = "";
  }
  if (url.startsWith("http") || url.startsWith("//") || !apiHost) {
    return url
  }
  return apiHost.host + url
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
