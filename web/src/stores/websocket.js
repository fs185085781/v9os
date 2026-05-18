import { reactive } from "vue";
import { defineStore } from "pinia";
import { getApiHost, uuid } from "@/util/util";
import emitter from "@/util/event.js";

export const websocketStore = defineStore("websocket", () => {
  const maxReconnectAttempts = 100;
  const cid = uuid();
  const clientMap = reactive({});

  function getListeners(client) {
    return Object.values(client.ons || {});
  }

  function buildSendFn(client) {
    return (msg) => {
      try {
        if (!client.ws || client.ws.readyState !== WebSocket.OPEN) {
          return false;
        }
        client.ws.send(msg);
        return true;
      } catch (e) {
        return false;
      }
    };
  }

  function emitOpen(client) {
    const sendFn = buildSendFn(client);
    getListeners(client).forEach((item) => {
      try {
        item.onOpen?.(sendFn);
      } catch (e) { }
    });
  }

  function emitMessage(client, data) {
    getListeners(client).forEach((item) => {
      try {
        item.onMessage?.(data);
      } catch (e) { }
    });
  }

  function emitClose(client, msg, readyState) {
    getListeners(client).forEach((item) => {
      try {
        item.onClose?.(msg, readyState);
      } catch (e) { }
    });
  }

  async function addClient(path, id, onOpen, onMessage, onClose, force) {
    if (!id) {
      throw new Error("websocket listener id is required");
    }
    let client = clientMap[path];
    if (!client) {
      client = {
        path,
        reconnectAttempts: 0,
        ons: {},
      };
      clientMap[path] = client;
    }
    client.ons[id] = { onOpen, onMessage, onClose };
    if (client.ws?.readyState === WebSocket.OPEN) {
      try {
        onOpen?.(buildSendFn(client));
      } catch (e) { }
      return id;
    }
    if (!force && (client.connecting || client.ws?.readyState === WebSocket.CONNECTING)) {
      return id;
    }
    await connectClient(client, force);
    return id;
  }

  async function connectClient(client, force) {
    if (!clientMap[client.path]) {
      return;
    }
    if (!force && (client.connecting || client.ws?.readyState === WebSocket.OPEN || client.ws?.readyState === WebSocket.CONNECTING)) {
      return;
    }
    client.connecting = true;
    const ah = await getApiHost();
    const wh = ah.replace("http", "ws");
    const wsUrl = `${wh}/api/ws${client.path}?cid=${cid}&token=${localStorage.getItem("token")}`;
    try {
      client.ws = new WebSocket(wsUrl);
      client.ws.onmessage = (event) => {
        emitMessage(client, event.data);
      };
      client.ws.onopen = () => {
        client.connecting = false;
        client.inReconnection = false;
        client.reconnectAttempts = 0;
        emitOpen(client);
      };
      client.ws.onclose = () => {
        client.connecting = false;
        handleReconnection(client.path, "Websocket连接被关闭");
      };
      client.ws.onerror = () => {
        client.connecting = false;
        handleReconnection(client.path, "Websocket连接出错");
      };
    } catch (error) {
      client.connecting = false;
      handleReconnection(client.path, "Websocket连接失败");
    }
  }

  function handleReconnection(path, msg) {
    const client = clientMap[path];
    if (!client) {
      return;
    }
    if (client.inReconnection) {
      return;
    }
    if (getListeners(client).length === 0) {
      removeClient(path);
      return;
    }
    client.inReconnection = true;
    if (client.reconnectAttempts >= maxReconnectAttempts) {
      emitClose(client, "超过最大连接失败次数", client.ws?.readyState);
      return;
    }
    emitClose(client, msg, client.ws?.readyState);
    client.reconnectAttempts++;
    setTimeout(() => {
      if (!clientMap[path]) {
        return;
      }
      client.inReconnection = false;
      connectClient(client, true);
    }, client.reconnectAttempts * 2000);
  }

  function removeClient(path, id) {
    const client = clientMap[path];
    if (!client) {
      return;
    }
    if (id) {
      delete client.ons[id];
      if (getListeners(client).length > 0) {
        return;
      }
    }
    delete clientMap[path];
    if (client.ws) {
      client.ws.close();
    }
  }

  function removeAllClients() {
    Object.keys(clientMap).forEach((path) => removeClient(path));
  }

  emitter.on("token-expired", removeAllClients);
  window.$websocket = { addClient, removeClient, removeAllClients };
  return window.$websocket;
});
