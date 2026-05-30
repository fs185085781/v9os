import mitt from "mitt";
import { onUnmounted } from "vue";

const emitter = mitt();

const isDebug = () => {
  try {
    return localStorage.getItem("v9os.debug.events") === "1";
  } catch {
    return false;
  }
};

const debugLog = (type, eventName, data, meta) => {
  if (!isDebug()) {
    return;
  }
  try {
    console.debug(`[v9os:event:${type}]`, eventName, data, meta || "");
  } catch {
    // ignore debug errors
  }
};

export function emitEvent(eventName, data, meta) {
  if (!eventName) {
    return;
  }
  debugLog("emit", eventName, data, meta);
  emitter.emit(eventName, data);
}

export function onEvent(eventName, callback) {
  if (!eventName || typeof callback !== "function") {
    return () => {};
  }
  debugLog("on", eventName);
  emitter.on(eventName, callback);
  return () => {
    debugLog("off", eventName);
    emitter.off(eventName, callback);
  };
}

export function offEvent(eventName, callback) {
  if (!eventName) {
    return;
  }
  debugLog("off", eventName);
  emitter.off(eventName, callback);
}

export function onceEvent(eventName, callback) {
  if (!eventName || typeof callback !== "function") {
    return () => {};
  }
  const off = onEvent(eventName, (data) => {
    off();
    callback(data);
  });
  return off;
}

window.addEventListener("message", (event) => {
  const msg = event.data;
  if (!msg || msg.__v9os !== true || msg.channel !== "plugin" || !msg.action) {
    return;
  }
  emitEvent(msg.action, msg.data, {
    source: "postMessage",
    v9os: true,
    channel: msg.channel,
    version: msg.version,
    origin: event.origin,
  });
});

export function useEventBus(eventName, callback) {
  const off = onEvent(eventName, callback);
  onUnmounted(() => {
    off();
  });
  return off;
}

export default emitter;
