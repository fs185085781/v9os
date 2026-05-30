# 前端插件开发

[返回总入口](starts.md) | [AI开发前端插件](ai-web-plugin-rules.md) | [主程序插件开发](plugin-main.md) | [第三方插件开发](plugin-third.md) | [疑难解答](troubleshooting.md)

前端插件是 `PluginType = 2` 的插件。它不需要独立后端进程，只要把前端文件放到运行目录的 `plugins/web/{code}` 中，并在内核插件表创建同编码记录即可。

参考目录：`../../plugins/web_demo`

## 一、适用场景

前端插件适合：

- 纯展示、报表、低代码页面。
- 调用已有 HTTP API 的工具页。
- 使用 `$v9os.api.webDataPost` 保存少量用户级插件数据。
- 需要快速接入 V9OS 窗口、主题、右键菜单和父子窗口通信的页面。

如果需要 Go 后端、数据库虚拟表、队列、锁、事件订阅等能力，优先使用 [主程序插件开发](plugin-main.md)。

## 二、目录结构

运行目录中结构如下：

```text
plugins/web/web_demo/
├── index.json
├── index.html
├── logo.png
├── child.html
└── v9os-web-demo.js
```

内核访问路径是：

```text
http://127.0.0.1:9099/api/webplugin/web_demo/
```

内核会检查 `index.json` 和 `index.html` 是否存在；如果不存在，会尝试从应用商店安装同编码插件包。

## 三、引入 SDK

前端插件入口需要引入：

```html
<script src="../../assets/sdk.js"></script>
```

注意：如果入口文件层级不同，请按实际访问路径调整相对层级。`plugins/web_demo/index.html` 当前示例中使用的是相对到 `assets/sdk.js` 的路径。

SDK 提供：

- 主题变量注入与 `window.onPersonalChange`。
- `$v9os.api.webDataPost(code, action, payload)`。
- `$v9os.event.on()` / `$v9os.event.emit()`。
- `$v9os.invoke()` 调用内核暴露对象。
- `$v9os.contextMenu` 右键菜单能力。
- `$v9os.msg.success()` / `$v9os.msg.error()` 消息提示。

## 四、创建插件数据

前端插件不需要启动调试进程。开发时直接把目录放入运行目录：

```text
运行目录/plugins/web/web_demo
```

然后在内核中创建同编码插件数据：

| 字段 | 示例 |
| --- | --- |
| `Code` | `web_demo` |
| `Name` | `前端插件演示` |
| `PluginType` | `2` |
| `Status` | `1` |
| `Version` | `1.0.1` |
| `IconUrl` | `/api/webplugin/web_demo/logo.png` |

访问路径：

```text
http://127.0.0.1:9099/api/webplugin/web_demo/
```

也可以通过应用商店模块打开，前端会按 `PluginType=2` 拼接 `/api/webplugin/{code}/`。

## 五、数据读写

前端插件的轻量持久化使用内核的插件扩展数据表。示例：

```js
await $v9os.api.webDataPost("web_demo", "set", {
  key: "demo_note",
  value: "hello"
});

const value = await $v9os.api.webDataPost("web_demo", "get", {
  key: "demo_note"
});

await $v9os.api.webDataPost("web_demo", "del", {
  key: "demo_note"
});
```

适合保存用户偏好、页面状态、简单草稿等小体量数据。复杂业务数据建议走主程序插件或内核业务接口。

## 六、窗口与主题

前端插件运行在内核窗口 iframe 中，窗口 ID 会由 SDK 通过 postMessage 获取。打开子窗口时建议通过内核窗口系统：

```js
$v9os.invoke("$wins", "addWindow", {
  width: 620,
  height: 430,
  title: "子窗口",
  iframeUrl: `${$v9os.host}/api/webplugin/web_demo/child.html`
}, window.__winId);
```

主题适配建议：

- 使用 `--user-primary-color` 作为主色。
- 使用 `--user-bg-1-color`、`--user-surface-color`、`--user-readable-surface-color` 作为背景。
- 使用 `--user-border-color` 作为边框。
- 使用 `calc(var(--user-round-enabled, 1) * 8px)` 控制圆角。
- 监听 `window.onPersonalChange` 更新语言、主题和渲染状态。

## 七、打包与安装

前端插件包至少包含：

```text
index.json
index.html
业务静态资源
```

`index.json` 可参考 `../../plugins/web_demo/index.json`：

```json
{
  "Name": "前端插件演示",
  "Description": "用于演示前端插件的数据读写删除、窗口打开关闭通讯，以及主题跟随变更。",
  "Code": "web_demo",
  "Version": "1.0.1",
  "PluginType": "2",
  "Category": "其他",
  "LimitVersion": "1.0.0",
  "IconUrl": "/api/webplugin/web_demo/logo.png",
  "Log": "首发"
}
```

## 八、常见问题

- 页面直接用浏览器打开本地文件时 `$v9os` 不存在：需要从内核 `/api/webplugin/{code}/` 打开。
- 主题不跟随：确认引入了 SDK，并监听或消费 SDK 写入的 CSS 变量。
- 插件打不开：确认 `plugins/web/{code}/index.html` 存在，插件表 `Code` 一致，`PluginType=2`，`Status=1`。

