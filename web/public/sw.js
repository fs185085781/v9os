importScripts("assets/simplehash.js");
const CACHE_NAME = "v9os-sw-cache";
let currentMode = "online";
let lastCheckTime = 0;
const notifyModeChange = () => {
  self.clients.matchAll().then((clients) => {
    clients.forEach((client) => {
      client.postMessage({
        action: "sw-change-status",
        data: { mode: currentMode },
      });
    });
  });
};
const generateCacheKey = async (req) => {
  let key = req.url;
  if (req.method != "GET") {
    try {
      const reqClone = req.clone();
      const body = await reqClone.blob();
      key += await simpleFileHash(body);
    } catch (error) {
      key += `|body:read-failed`;
    }
  }
  return key;
};
const cleanExpiredCaches = async (maxAge = 60 * 24 * 60 * 60 * 1000) => {
  const cache = await caches.open(CACHE_NAME);
  const requests = await cache.keys();
  const now = Date.now();
  for (const request of requests) {
    const response = await cache.match(request);
    if (!response) continue;
    const cacheTime = response.headers.get("cache-time");
    if (cacheTime && now - cacheTime > maxAge) {
      await cache.delete(request);
    }
  }
};
self.addEventListener("install", (event) => {
  event.waitUntil(self.skipWaiting());
});
self.addEventListener("activate", (event) => {
  Promise.all([cleanExpiredCaches()]).then(() => self.clients.claim());
});
self.addEventListener("message", async (event) => {
  const param = event.data;
  if (!param) {
    return;
  }
  if (param.type === "GET_MODE") {
    event.ports[0].postMessage({
      data: currentMode,
      qid: param.qid,
    });
  }
});
const noErrAction = async (fn) => {
  try {
    return await fn();
  } catch (error) {
    console.error(error);
  }
};
const noErrFetch = async (req, options = undefined) => {
  if (options && options.redirect == "manual" && req.mode == "no-cors") {
    options = undefined;
  }
  return await noErrAction(() => fetch(req, options));
};
self.addEventListener("fetch", (event) => {
  const request = event.request;
  event.respondWith(
    (async () => {
      let key = await generateCacheKey(request);
      if (!key.startsWith("http")) {
        return await noErrFetch(request);
      }
      if (currentMode == "online") {
        const networkResponse = await noErrFetch(request, {
          redirect: "manual",
        });
        if (networkResponse) {
          const status = networkResponse.status;
          if (status >= 200 && status < 400 && currentMode != "online") {
            currentMode = "online";
            notifyModeChange();
          }
          if (networkResponse.ok && status != 206) {
            const contentLength = networkResponse.headers.get("content-length");
            if (!(contentLength && contentLength > 10485760)) {
              await noErrAction(async () => {
                const cache = await caches.open(CACHE_NAME);
                const response = await networkResponse.clone();
                const modifiedHeaders = new Headers(response.headers);
                modifiedHeaders.set("cache-time", Date.now());
                cache.put(
                  key,
                  new Response(response.body, {
                    status: response.status,
                    statusText: response.statusText,
                    headers: modifiedHeaders,
                  }),
                );
              });
            }
          }
          if (status >= 300 && status < 400) {
            const location = networkResponse.headers.get("Location");
            if (location) {
              return await noErrFetch(location, {
                redirect: "follow",
                credentials: "include",
              });
            }
          }
          return networkResponse;
        }
      }
      if (currentMode == "offline" && Date.now() - lastCheckTime > 10000) {
        lastCheckTime = Date.now();
        try {
          fetch(request, { redirect: "manual" }).then((networkResponse) => {
            const status = networkResponse.status;
            if (status >= 200 && status < 400) {
              currentMode = "online";
              notifyModeChange();
            }
          });
        } catch (error) {}
      }
      const cachedResponse = await caches.match(key);
      if (cachedResponse) {
        if (currentMode != "offline") {
          currentMode = "offline";
          notifyModeChange();
        }
        return cachedResponse;
      }
      return new Response(
        `{"code":-1,"msg":"当前网络不佳,请尽快检查并恢复在线模式"}`,
        {
          status: 200,
          statusText: "Not Cached",
        },
      );
    })(),
  );
});
