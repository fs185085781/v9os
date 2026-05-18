import { reactive, nextTick } from 'vue'
import { defineStore } from 'pinia'
import { uuid, postData } from "@/util/util"
import { useEventBus } from '@/util/event.js';
import emitter from '@/util/event.js';
export const windowsStore = defineStore('windows', () => {
  const windows = reactive([])
  const common = reactive({
    inDraggable: false,
    zIndex: 10
  })
  const raiseWindow = (win) => {
    if (!win) {
      return
    }
    common.zIndex += 2
    win.zIndex = common.zIndex
  }
  const getOpenWindow = (id) => windows.filter(x => x.id == id && !x.close).shift()
  const getWindowStackChain = (win) => {
    const chain = []
    let parent = win?.parentId ? getOpenWindow(win.parentId) : null
    while (parent) {
      chain.unshift(parent)
      parent = parent.parentId ? getOpenWindow(parent.parentId) : null
    }
    if (win) {
      chain.push(win)
    }
    let child = win ? windows.filter(x => x.parentId == win.id && !x.close).shift() : null
    while (child) {
      chain.push(child)
      child = windows.filter(x => x.parentId == child.id && !x.close).shift()
    }
    return chain
  }
  const raiseWindowChain = (win) => {
    getWindowStackChain(win).forEach(raiseWindow)
  }
  const getVisibleWindowChain = (win) => {
    return getWindowStackChain(win).filter(x => x && !x.close && x.status != "min")
  }
  const isWindowChainOnTop = (win) => {
    const chain = getVisibleWindowChain(win)
    if (!chain.length) {
      return true
    }
    for (let i = 1; i < chain.length; i++) {
      if ((chain[i].zIndex || 0) <= (chain[i - 1].zIndex || 0)) {
        return false
      }
    }
    const baseZIndex = chain[0].zIndex || 0
    return windows
      .filter(x => x && !x.close && x.status != "min" && !chain.includes(x))
      .every(x => (x.zIndex || 0) < baseZIndex)
  }
  const ensureWindowChainOnTop = (win) => {
    if (!isWindowChainOnTop(win)) {
      raiseWindowChain(win)
    }
  }
  const addWindow = function (win, parentId) {
    
    //如果parentId不为空,则是子窗口,子窗口不支持最小化,不支持最大化,不支持调节尺寸
    let fyWin = null;
    if (win.id) {
      fyWin = windows.filter(x => { return !x.close && x.id == win.id }).shift()
    }
    if(!fyWin){
      fyWin = windows.filter(x => { x.active = false; return x.close }).shift()
    }
    if (fyWin) {
      const fyId = fyWin.id
      const fyWinKeys = Object.keys(fyWin);
      const winKeys = Object.keys(win);
      for (const key of winKeys) {
        Object.defineProperty(fyWin, key, {
          value: win[key],
          writable: true,
          enumerable: true,
          configurable: true
        });
      }
      for (const key of fyWinKeys) {
        if (!winKeys.includes(key)) {
          Object.defineProperty(fyWin, key, {
            value: null,
            writable: true,
            enumerable: true,
            configurable: true
          });
        }
      }
      fyWin.id = fyId
      if (fyWin.iframeUrl) {
        const tmpUrl = fyWin.iframeUrl
        fyWin.iframeUrl = ""
        nextTick().then(() => {
          fyWin.iframeUrl = tmpUrl
        });
      }
    } else {
      fyWin = win
      fyWin.id = win.id || uuid()
      windows.push(fyWin)
    }
    fyWin.close = false
    fyWin.active = true
    fyWin.status = "normal"
    fyWin.lastStatus = fyWin.status
    fyWin.parentId = parentId
    let parentWin = null
    if (parentId) {
      parentWin = windows.filter(x => x.id == parentId && !x.close).shift()
    }
    if (parentWin) {
      fyWin.left = parentWin.left + (parentWin.width - fyWin.width) / 2
      fyWin.top = parentWin.top + (parentWin.height - fyWin.height) / 2
      common.parentId = parentWin.id
    } else {
      fyWin.left = (document.body.clientWidth - fyWin.width) / 2
      fyWin.top = (document.body.clientHeight - 24 - fyWin.height) / 2
    }
    raiseWindowChain(fyWin)
  }
  const activeWindow = (win) => {
    const fyWin = windows.filter(x => { x.active = false; return x == win }).shift()
    if (!fyWin) {
      return false
    }
    let hasChild = false
    let child = windows.filter(x => x.parentId == fyWin.id && !x.close).shift()
    while (child) {
      hasChild = true
      const tmp = windows.filter(x => x.parentId == child.id && !x.close).shift()
      if (tmp) {
        child = tmp
      } else {
        break
      }
    }
    if (!hasChild) {
      fyWin.active = true
      common.parentId = fyWin.parentId
      ensureWindowChainOnTop(fyWin)
      return true;
    }
    child.active = true
    common.parentId = fyWin.id
    ensureWindowChainOnTop(child)
    child.inBlink = true
    setTimeout(() => {
      child.inBlink = false
    }, 300);
    return false
  }
  const closeWindow = (winId) => {
    const fyWin = windows.filter(x => x.id == winId && !x.close).shift()
    if (!fyWin) {
      return;
    }
    fyWin.close = true
    fyWin.iframeUrl = ""
    fyWin.component = null
    if (fyWin.parentId) {
      const parent = windows.filter(x => x.id == fyWin.parentId && !x.close).shift()
      if (parent) {
        activeWindow(parent)
        return
      }
    }
    const tmp = windows.filter(x => !x.parentId && !x.close).shift()
    if (tmp) {
      activeWindow(tmp)
    }
  }
  const getMaxWindow = () => {
    const data = windows.filter(x => x.status == "max" && !x.close).shift()
    return data
  }
  const initPostMessage = (iframeUi, winId) => {
    if (!window.__v9osIframes) {
      window.__v9osIframes = {}
    }
    window.__v9osIframes[winId] = iframeUi
    if (!window._pluginPostData) {
      window._pluginPostData = async function (module, action, param) {
        return await postData('../plugin', module + '/' + action, param, 'json')
      }
    }
    useEventBus("win-id-get", (msg) => {
      const iframeWindow = iframeUi.value?.contentWindow
      if (!iframeWindow) {
        return
      }
      iframeWindow.postMessage({ action: "win-id-get", data: { id: winId } }, "*");
    })
    useEventBus("iframe-invoke", async (msg) => {
      if (msg.winId != winId) {
        return
      }
      const iframeWindow = iframeUi.value?.contentWindow
      if (!iframeWindow) {
        return
      }
      try {
        const obj = getValueByPathAdvanced(window, msg.entity)
        let res
        if (typeof obj[msg.method] == "function") {
          res = await obj[msg.method](...msg.param)
        } else {
          res = await obj[msg.method]
        }
        res = JSON.parse(JSON.stringify(res));
        iframeWindow.postMessage({ action: "iframe-invoke", data: { taskId: msg.taskId, res, code: 0 } }, "*");
      } catch (err) {
        iframeWindow.postMessage({ action: "iframe-invoke", data: { taskId: msg.taskId, msg: err.message, code: -1 } }, "*");
      }
    })
    useEventBus("iframe-event-on", async (msg) => {
      if (msg.winId != winId) {
        return
      }
      const iframeWindow = iframeUi.value?.contentWindow
      if (!iframeWindow) {
        return
      }
      useEventBus(msg.eventName, (data) => {
        iframeWindow.postMessage({ action: "iframe-event-on", data: { eventName: msg.eventName, data } }, "*");
      })
      if (msg.eventName == "personalChange") {
        const settings = $user.settings
        emitter.emit(msg.eventName, { "Theme": settings.Theme, "Color": settings.Color, "Round": settings.Round, "Lang": settings.Lang, "Mode": settings.Mode, "Font": settings.Font, "ColorDesc": settings.ColorDesc })
      }
    })
    const getValueByPathAdvanced = (targetObj, path) => {
      if (typeof targetObj !== 'object' || targetObj === null) {
        return undefined;
      }
      if (typeof path !== 'string' || path === '') {
        return targetObj;
      }
      const regex = /\.?([^.[\]]+)|\[(\d+)\]|\["([^"]+)"\]|\['([^']+)'\]/g;
      const parts = [];
      let match;
      while ((match = regex.exec(path)) !== null) {
        const part = match[1] || match[2] || match[3] || match[4];
        if (part) {
          parts.push(part);
        }
      }
      let current = targetObj;
      for (const part of parts) {
        if (current == null) {
          return undefined;
        }
        current = current[part];
      }
      return current;
    }
  }
  const winStatusChange = (winData, status, tapOnWindow) => {
    if (winData.parentId && (status == "max" || status == "min")) {
      return
    }
    let child = windows.filter(x => x.parentId == winData.id && !x.close).shift()
    if (child) {
      return
    }
    if (winData.status == status) {
      return
    }
    winData.lastStatus = winData.status
    if (status == "close") {
      closeWindow(winData.id)
    } else {
      if (status == "max" && winData.lastStatus == "normal") {
        winData.nwidth = winData.width
        winData.nheight = winData.height
        winData.nleft = winData.left
        winData.ntop = winData.top
      }
      if (status == "max") {
        winData.status = "max"
        winData.width = document.body.clientWidth
        winData.height = document.body.clientHeight
        winData.top = 0
        winData.left = 0
      } else if (status == "normal") {
        if (winData.lastStatus == "max") {
          winData.width = winData.nwidth
          winData.height = winData.nheight
          winData.top = winData.ntop
          winData.left = winData.nleft
        }
        winData.status = "normal"
        if (tapOnWindow) {
          tapOnWindow(false);
        }
      } else if (status == "min") {
        winData.status = "min"
      }
    }
    emitter.emit("window-status-change", {});
  }
  window.$wins = { windows, common, addWindow, activeWindow, closeWindow, getMaxWindow, initPostMessage, winStatusChange, raiseWindow, raiseWindowChain }
  return window.$wins
})
