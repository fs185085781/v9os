# AI 前端插件开发规则（PluginType=2）
[返回总入口](starts.md) | [前端插件开发](plugin-web.md)

> 用途：把本文件和业务需求交给 AI，即可生成 V9OS 前端插件代码。AI 必须优先遵守本文件。

## 1. 适用范围

前端插件是 `PluginType = 2` 的插件。它没有独立 Go 后端进程，只需要把静态前端文件放到运行目录 `plugins/web/{code}` 中，并在内核插件表创建同编码记录。

以下场景优先使用前端插件：

- 纯展示页、仪表盘、报表、低代码页面。
- 调用已有 HTTP API 的工具页。
- 使用 `$v9os.api.webDataPost` 保存少量用户级插件数据，例如偏好、草稿、页面状态。
- 需要快速接入 V9OS 窗口、主题、右键菜单、父子窗口通信。
- 不需要 Go 后端、数据库模型、事务、后台任务、分布式锁或事件订阅后端。

如果业务需要结构化数据库、复杂 CRUD、后台任务、队列、锁、事件订阅、对接内核后端能力，必须改用主程序插件规则。

## 2. AI 开发前端插件的固定流程

AI 收到业务需求后必须按此顺序执行：

1. **确定插件编码**：使用小写蛇形命名，例如 `kanban_board`。编码必须用于目录名、`index.json.Code`、图标路径、`api.webDataPost` 模块名、页面 URL。
2. **确认插件类型**：无独立后端时使用 `PluginType=2`。
3. **选择前端形态**：默认优先原生 HTML/CSS/JS，除非用户要求 Vue/Vite 或业务复杂到需要组件化。
4. **生成静态目录**：至少包含 `index.json`、`index.html`、图标和业务 JS/CSS。
5. **引入 SDK**：入口页面必须引入 `assets/sdk.js`，路径按访问层级计算。
6. **封装 SDK 调用**：数据读写、窗口、事件、右键菜单、主题变化统一封装，业务代码不要散落硬编码。
7. **适配主题**：CSS 必须使用 V9OS 主题变量。
8. **自检访问路径**：确认能从 `/api/webplugin/{code}/` 打开。
9. **输出交付说明**：列出文件、访问路径、插件表记录、主要能力。

禁止：把前端插件写成需要 Go 后端的结构；把大量业务数据塞进 `api.webDataPost`；本地 file 直接打开后假设 `$v9os` 存在；复制 `web_demo` 后漏改编码；把所有前端逻辑、样式和状态都揉进一个 `app.js` 或一个超大 `index.html`。

## 3. 标准目录结构

开发目录推荐：

```text
plugins/{code}/
├── index.json
├── index.html
├── logo.png 或 logo.svg
├── css/
│   ├── theme.css
│   ├── layout.css
│   └── 业务域样式.css
├── js/
│   ├── main.js          # 入口，只做初始化和路由挂载
│   ├── api.js           # SDK/API 封装
│   ├── state.js         # 状态
│   ├── layout.js        # 布局/导航
│   ├── project.js       # 项目领域示例；按业务替换
│   ├── task.js          # 任务领域示例；按业务替换
│   ├── modal.js         # 弹窗/抽屉
│   └── util.js          # 通用工具
└── 可选：child.html、assets/
```

运行目录中通常为：

```text
plugins/web/{code}/
├── index.json
├── index.html
├── logo.png
└── 业务静态资源
```

内核访问路径：

```text
http://127.0.0.1:9099/api/webplugin/{code}/
```

如果 `index.json` 或 `index.html` 不存在，内核可能尝试从应用商店安装同编码插件包。

## 4. 前端代码组织规则

无论是前端插件还是主程序插件附带页面，前端代码都必须像可交付项目一样分层组织。

强制规则：

- **入口只做启动**：`js/main.js` 只负责初始化主题、拉取首屏数据、绑定路由，不写具体业务大段逻辑。
- **API 单独封装**：所有 `$v9os.api.webDataPost`、`$v9os.api.pluginPost`、`$v9os.invoke` 调用必须放入 `js/api.js` 或按领域拆分的 API 文件。
- **状态单独管理**：共享状态放入 `js/state.js`，领域内状态可放在对应领域文件，不要散落全局变量。
- **每个业务域一个文件**：例如项目、任务、看板、列表、日历、甘特、后台、设置、评论、附件等分别独立文件。
- **每个复杂视图一个文件**：看板、日历、甘特、统计图、任务详情抽屉不得全部写在同一个文件。
- **通用 UI 单独文件**：弹窗、抽屉、右键菜单、消息封装放入 `modal.js`、`context_menu.js` 等。
- **工具函数单独文件**：日期、转义、格式化、DOM 帮助函数放入 `util.js`。
- **CSS 按层级拆分**：主题变量、布局、组件、领域样式分文件；禁止把复杂业务全部写入一个巨大 `<style>`。
- **文件大小约束**：普通 JS 文件建议不超过 400 行；超过时必须继续按子域拆分。
- **模块加载方式**：原生前端优先使用 `<script type="module">` 和 ES module；无需构建也能保持清晰结构。

推荐结构：

```text
static/
├── index.html
├── css/
│   ├── theme.css
│   ├── layout.css
│   ├── components.css
│   └── board.css
└── js/
    ├── main.js
    ├── api.js
    ├── state.js
    ├── util.js
    ├── layout.js
    ├── project.js
    ├── task.js
    ├── board.js
    ├── list_view.js
    ├── calendar_view.js
    ├── gantt_view.js
    ├── analysis_view.js
    ├── modal.js
    └── context_menu.js
```

## 5. `index.json` 规则

前端插件必须提供 `index.json`：

```json
{
  "Name": "看板工具",
  "Description": "轻量看板和个人任务视图",
  "Code": "kanban_board",
  "Version": "1.0.0",
  "PluginType": "2",
  "Category": "效率工具",
  "LimitVersion": "1.0.0",
  "IconUrl": "/api/webplugin/kanban_board/logo.png",
  "Log": "首发"
}
```

强制规则：

- `Code` 必须等于目录名和前端常量 `pluginCode`。
- `PluginType` 必须为字符串 `"2"`。
- `IconUrl` 前端插件使用 `/api/webplugin/{code}/logo.png` 或 `/api/webplugin/{code}/logo.svg`。
- `Version` 从 `1.0.0` 起步。

## 5. SDK 引入规则

前端插件入口必须引入 SDK：

```html
<script src="../../../assets/sdk.js"></script>
```

路径说明：

- 当前示例 `plugins/web_demo/index.html` 使用 `../../../assets/sdk.js`。
- 主程序插件页面 `/page/{code}/` 常用 `../../assets/sdk.js`。
- 前端插件页面从 `/api/webplugin/{code}/` 访问时，相对层级必须能定位到内核 `assets/sdk.js`。
- 如果页面层级更深，例如 `pages/detail.html`，必须按实际 URL 调整为 `../../../../assets/sdk.js` 或使用正确相对路径。

SDK 会注入 `window.$v9os`，常用能力：

```js
$v9os.api.webDataPost(code, action, payload, showType)
$v9os.api.pluginPost(module, action, payload, showType)
$v9os.invoke(entity, method, ...param)
$v9os.event.on(eventName, callback)
$v9os.event.off(eventName, callback)
$v9os.event.emit(eventName, data)
$v9os.contextMenu.show(options)
$v9os.contextMenu.onAction(actionName, callback)
$v9os.msg.success(message)
$v9os.msg.error(message)
$v9os.msg.alert(content, title)
$v9os.msg.confirm(content, title)
$v9os.msg.prompt(content, title)
$v9os.file.selectFile(relative, title, ext, save)
$v9os.file.selectLongDir(relative, title)
$v9os.file.saveFile(title, name, blob)
$v9os.theme.apply(settings)
```

## 6. 运行环境规则

- 插件必须从内核 `/api/webplugin/{code}/` 打开。
- 直接双击本地 HTML 时 `$v9os` 不存在；可做本地预览降级，但正式功能必须依赖 SDK。
- 需要当前窗口 ID 时使用 `window.__winId`；它由 SDK 通过 postMessage 自动获取。
- SDK 初始化后会设置 `$v9os.host`，可用于拼接内核 URL。

推荐写法：

```js
const pluginCode = "kanban_board";

function hasSdk() {
  return !!window.$v9os;
}

function ensureSdk() {
  if (!hasSdk()) throw new Error("请从 V9OS /api/webplugin/kanban_board/ 打开插件");
}
```

## 7. 轻量数据读写规则

前端插件只适合保存用户级轻量键值数据。内核实际按 `user_id + code + data_key` 隔离。

SDK 调用：

```js
await $v9os.api.webDataPost("kanban_board", "set", {
  key: "board_state",
  val: JSON.stringify(state)
});

const value = await $v9os.api.webDataPost("kanban_board", "get", {
  key: "board_state"
});

await $v9os.api.webDataPost("kanban_board", "del", {
  key: "board_state"
});
```

注意：底层接口读取的是 `key` 和 `val`，推荐 `set` 时传 `{ key, val }`。如果历史代码使用 `{ key, value }`，AI 新代码必须优先改为 `{ key, val }`。

封装模板：

```js
const storage = {
  async set(key, value) {
    return await $v9os.api.webDataPost(pluginCode, "set", {
      key,
      val: typeof value === "string" ? value : JSON.stringify(value)
    }, "err");
  },
  async get(key, fallback = null) {
    const raw = await $v9os.api.webDataPost(pluginCode, "get", { key }, "err");
    if (raw === false || raw == null || raw === "") return fallback;
    try { return JSON.parse(raw); } catch { return raw; }
  },
  async del(key) {
    return await $v9os.api.webDataPost(pluginCode, "del", { key }, "err");
  }
};
```

使用限制：

- 适合偏好设置、页面布局、草稿、小型 JSON。
- 不适合项目/任务/订单等复杂业务数据库。
- 不适合大文件、二进制、频繁写入日志。
- 复杂业务必须改用主程序插件。

## 8. 调用已有接口规则

如果前端插件调用已有 HTTP API：

- 优先使用系统已经暴露的 API 或 `$v9os.invoke` 能力。
- 不要把敏感 token 写死在前端。
- 对第三方跨域 API，要确认 CORS；不满足时应改用主程序插件后端代理。
- 所有用户输入进入 URL 或请求体前要做校验、编码和错误提示。

前端插件不应调用不存在的主程序插件动作。如果业务需要 `$v9os.api.pluginPost("某插件", action)`，必须说明该插件已安装并提供该动作。

## 9. 窗口和子窗口规则

打开子窗口必须通过内核窗口系统：

```js
$v9os.invoke("$wins", "addWindow", {
  width: 620,
  height: 430,
  title: "子窗口",
  iframeUrl: `${$v9os.host}/api/webplugin/kanban_board/child.html`
}, window.__winId);
```

父子通信推荐事件模式：

```js
// 父窗口
$v9os.event.on("kanbanChild", async (data) => {
  if (data?.action === "get") {
    $v9os.event.emit("kanbanChild-win", { action: "data", board: currentBoard });
  }
  if (data?.action === "refresh") {
    if (data.winId) $v9os.invoke("$wins", "closeWindow", data.winId);
    await reload();
  }
});

// 子窗口
$v9os.event.on("kanbanChild-win", (data) => {
  console.log("parent data", data);
});
$v9os.event.emit("kanbanChild", { action: "get", winId: window.__winId });
```

规则：

- 事件名必须加插件前缀，避免冲突，例如 `kanbanChild`、`kanbanChild-win`。
- 页面卸载时调用 `event.off` 清理监听。
- 子窗口关闭并通知父窗口时，带上 `winId: window.__winId`，父窗口负责关闭子窗口并刷新。

## 10. 右键菜单规则

使用内核右键菜单：

```js
$v9os.contextMenu.onAction("refreshKanban", () => reload());

panel.addEventListener("contextmenu", (event) => {
  event.preventDefault();
  $v9os.contextMenu.show({
    x: event.clientX,
    y: event.clientY,
    type: "kanban_board.panel",
    payload: { pluginCode },
    items: [
      { key: "refresh", label: "刷新", group: "main", order: 10, actionId: "refreshKanban" }
    ]
  });
});
```

规则：

- `actionId` 必须与 `onAction` 注册的名称一致。
- 菜单项的 `key` 在当前菜单内唯一。
- 鼠标/触摸其他位置时 SDK 会尝试关闭菜单，业务无需重复实现。

## 11. 主题适配规则

前端插件必须使用 V9OS 主题变量，不要写死大面积颜色。

推荐 CSS：

```css
:root {
  --app-primary: var(--user-primary-color, #2080f0);
  --app-primary-hover: var(--user-primary-color-hover, #4098fc);
  --app-primary-text: var(--user-primary-text-color, #fff);
  --app-bg: var(--user-bg-1-color, #f6f8fb);
  --app-surface: var(--user-readable-surface-color, #fff);
  --app-surface-2: var(--user-surface-color, #fff);
  --app-control: var(--user-control-color, rgba(0,0,0,.04));
  --app-control-hover: var(--user-control-hover-color, rgba(0,0,0,.07));
  --app-border: var(--user-border-color, rgba(0,0,0,.13));
  --app-text: var(--user-text-1-color, #172033);
  --app-muted: var(--user-text-2-color, #667085);
  --app-radius: calc(var(--user-round-enabled, 1) * 8px);
}

html[data-theme="dark"] {
  color-scheme: dark;
}
```

监听个性化变化：

```js
window.onPersonalChange = (settings, theme) => {
  if (window.$v9os?.theme) $v9os.theme.apply(settings);
  document.documentElement.dataset.theme = settings.Theme === "dark" ? "dark" : "light";
  document.documentElement.dataset.lang = settings.Lang || "zh";
};

if (window.__personal && window.$v9os?.theme) {
  $v9os.theme.apply(window.__personal);
}
```

规则：

- 背景用 `--user-bg-1-color` 或 `--user-readable-surface-color`。
- 边框用 `--user-border-color`。
- 文本用 `--user-text-1-color`、`--user-text-2-color`、`--user-text-3-color`。
- 主按钮用 `--user-primary-color` 和 `--user-primary-text-color`。
- 圆角用 `calc(var(--user-round-enabled, 1) * Npx)`。
- 如果支持字体，字体文件路径可用 `${$v9os.host}/api/fonts/{font}.font`。

## 12. 消息、确认框和输入框规则

优先使用 SDK 消息能力：

```js
$v9os.msg.success("保存成功");
$v9os.msg.error("保存失败");

const ok = await $v9os.msg.confirm("确认删除吗？", "提示");
const name = await $v9os.msg.prompt("请输入名称", "新建");
await $v9os.msg.alert("操作完成", "提示");
```

规则：

- 删除、覆盖、清空等危险操作必须确认。
- 错误消息要展示给用户，不要只写 `console.error`。
- 本地预览模式可以降级为 `alert/confirm/prompt`。

## 13. 文件选择和保存规则

SDK 提供文件能力：

```js
const file = await $v9os.file.selectFile("false", "请选择文件", "png,jpg", "false");
const dir = await $v9os.file.selectLongDir("false", "请选择目录");
await $v9os.file.saveFile("保存文件", "export.json", new Blob([json], { type: "application/json" }));
```

规则：

- 文件选择依赖 `file_system` 插件页面。
- `relative` 只有字符串 `"true"` 会被当作 true，否则传 `"false"`。
- `ext` 用逗号分隔，例如 `png,jpg,pdf`。
- 保存文件时先让用户选择保存位置，再通过 SDK 返回的上传信息写入。

## 14. 原生 HTML/JS 插件模板

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>看板工具</title>
  <script src="../../../assets/sdk.js"></script>
  <style>
    :root {
      --app-primary: var(--user-primary-color, #2080f0);
      --app-primary-text: var(--user-primary-text-color, #fff);
      --app-bg: var(--user-bg-1-color, #f6f8fb);
      --app-surface: var(--user-readable-surface-color, #fff);
      --app-border: var(--user-border-color, rgba(0,0,0,.13));
      --app-text: var(--user-text-1-color, #172033);
      --app-muted: var(--user-text-2-color, #667085);
      --app-radius: calc(var(--user-round-enabled, 1) * 8px);
    }
    body { margin: 0; background: var(--app-bg); color: var(--app-text); font: 14px/1.6 system-ui, sans-serif; }
    .panel { margin: 16px; padding: 16px; border: 1px solid var(--app-border); border-radius: var(--app-radius); background: var(--app-surface); }
    button { border: 0; border-radius: var(--app-radius); background: var(--app-primary); color: var(--app-primary-text); padding: 8px 12px; cursor: pointer; }
  </style>
</head>
<body>
  <main class="panel">
    <h1>看板工具</h1>
    <textarea id="note" placeholder="输入草稿"></textarea>
    <button id="save">保存</button>
    <button id="load">读取</button>
    <button id="remove">删除</button>
    <pre id="output"></pre>
  </main>
  <script>
    const pluginCode = "kanban_board";
    const key = "draft";

    async function setData(value) {
      return await $v9os.api.webDataPost(pluginCode, "set", { key, val: value }, "okerr");
    }
    async function getData() {
      return await $v9os.api.webDataPost(pluginCode, "get", { key }, "err");
    }
    async function delData() {
      return await $v9os.api.webDataPost(pluginCode, "del", { key }, "okerr");
    }

    window.onPersonalChange = (settings) => {
      document.documentElement.dataset.theme = settings.Theme === "dark" ? "dark" : "light";
      document.documentElement.dataset.lang = settings.Lang || "zh";
    };

    document.querySelector("#save").onclick = async () => {
      await setData(document.querySelector("#note").value);
    };
    document.querySelector("#load").onclick = async () => {
      const value = await getData();
      document.querySelector("#output").textContent = value || "";
    };
    document.querySelector("#remove").onclick = async () => {
      if (await $v9os.msg.confirm("确认删除草稿吗？")) await delData();
    };
  </script>
</body>
</html>
```

## 15. 本地预览降级规则（可选）

为了方便静态预览，可以做降级，但正式运行必须使用 `$v9os`：

```js
function localKey(key) {
  return `${pluginCode}:${key}`;
}

const dataApi = {
  async set(key, value) {
    if (window.$v9os) return $v9os.api.webDataPost(pluginCode, "set", { key, val: value }, "err");
    localStorage.setItem(localKey(key), value);
    return true;
  },
  async get(key) {
    if (window.$v9os) return $v9os.api.webDataPost(pluginCode, "get", { key }, "err");
    return localStorage.getItem(localKey(key));
  },
  async del(key) {
    if (window.$v9os) return $v9os.api.webDataPost(pluginCode, "del", { key }, "err");
    localStorage.removeItem(localKey(key));
    return true;
  }
};
```

规则：

- 降级只用于 UI 预览，不能替代真实插件数据。
- 页面应提示“本地预览模式”或“SDK 未连接”。

## 16. 构建型前端规则（可选）

如使用 Vite/Vue/React：

- 构建产物必须能作为纯静态文件运行。
- `base` 构建时使用 `./`，保证 `/api/webplugin/{code}/` 下资源路径正确。
- `index.html` 仍必须引入正确的 `assets/sdk.js`。
- 不要在前端项目里引入后端私密配置。
- 生产包至少包含 `index.json`、`index.html`、构建产物、图标。

Vite 示例：

```js
export default defineConfig(({ command }) => ({
  base: command === "serve" ? "/api/webplugin/kanban_board/" : "./",
  build: { outDir: "../dist" }
}));
```

## 17. 权限和插件表规则

开发/安装前，内核插件表需要同编码记录：

| 字段 | 示例 |
| --- | --- |
| `Code` | `kanban_board` |
| `Name` | `看板工具` |
| `PluginType` | `2` |
| `Status` | `1` |
| `Version` | `1.0.0` |
| `IconUrl` | `/api/webplugin/kanban_board/logo.png` |

`api.webDataPost` 底层会检查当前用户对 `{code}/{method}` 是否有权限，其中 method 是 `get`、`set`、`del`。如果返回 403，需要在插件/权限系统里授权对应动作。

## 18. 常见业务拆分建议

适合前端插件的业务拆法：

- 偏好设置：一个 key 保存 JSON。
- 个人草稿：按 `draft:{id}` 保存小文本。
- 仪表盘配置：`dashboard_layout` 保存布局 JSON。
- 快捷入口：`shortcuts` 保存链接数组。
- 查询工具：表单参数存在页面状态，实际数据来自已有 API。

不适合前端插件的业务：

- 多人协作项目管理。
- 大量任务、订单、工单、客户数据。
- 需要权限范围的数据表。
- 需要后台同步、定时扫描、消息订阅。
- 需要隐藏密钥访问第三方 API。

这些必须使用主程序插件。

## 19. 错误处理规则

- 所有 SDK 调用都要处理失败返回；`api.pluginPost/api.webDataPost` 在非 `json` 模式下可能返回 `false`。
- 保存前校验必填字段。
- 删除前确认。
- JSON 解析必须 try/catch。
- 网络/权限错误要用 `$v9os.msg.error` 或页面错误态提示。
- 不要让按钮在重复点击时并发写入；可设置 loading 状态。

## 20. 生成后验收清单

AI 完成前端插件后必须自检：

- [ ] `index.json.Code`、目录名、前端 `pluginCode` 完全一致。
- [ ] `index.json.PluginType` 为 `"2"`。
- [ ] `IconUrl` 使用 `/api/webplugin/{code}/logo.png` 或 `.svg`。
- [ ] `index.html` 存在且引入正确层级的 `assets/sdk.js`。
- [ ] 页面从 `/api/webplugin/{code}/` 打开时资源路径正确。
- [ ] 数据读写使用 `$v9os.api.webDataPost(code, "get/set/del", ...)`。
- [ ] `set` 参数使用 `{ key, val }`。
- [ ] 复杂业务数据没有塞进前端键值存储。
- [ ] 窗口打开使用 `$v9os.invoke("$wins", "addWindow", ...)`。
- [ ] 父子通信使用 `$v9os.event.on/event.emit/event.off`，并清理监听。
- [ ] 右键菜单 `actionId` 和 `onAction` 一致。
- [ ] CSS 使用 V9OS 主题变量，支持深浅色、主色、圆角。
- [ ] 危险操作有确认框，错误有提示。
- [ ] 如支持本地预览，已明确区分预览数据和正式插件数据。
