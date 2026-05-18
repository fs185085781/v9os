# 主程序插件开发

[返回总入口](starts.md) | [前端插件开发](plugin-web.md) | [第三方插件开发](plugin-third.md) | [架构原理](architecture.md)

主程序插件是 `PluginType = 1` 的插件。它有独立 Go 进程，可以通过 `share/plugin` 调用内核能力，同时可以附带前端页面，通过 `/page/{code}/` 被主程序打开。

参考目录：`../../plugins/main_demo`

## 一、目录结构

典型结构如下：

```text
plugins/main_demo/
├── main.go
├── index.json
├── logo.png
├── impl/
│   └── *.go
├── static/
│   └── 构建后的前端静态文件
└── web/
    └── 插件前端源码
```

`main.go` 中通过 `plugin.Server("main_demo", staticFiles)` 启动插件服务。`impl` 包通过空导入注册插件后端动作。

## 二、前端 SDK

主程序插件前端需要引入内核 SDK：

```html
<script src="../../assets/sdk.js"></script>
```

`plugins/main_demo/web/index.html` 已按这个方式引入。SDK 会注入 `$v9os`，常用能力包括：

- `$v9os.pluginPost(code, action, payload)`：调用主程序插件后端动作。
- `$v9os.eventOn()` / `$v9os.eventEmit()`：窗口和插件页面间事件通信。
- `$v9os.invoke("$wins", "addWindow", options, parentWinId)`：请求内核打开窗口。
- `$v9os.contextMenu.show()`：请求内核渲染右键菜单。
- `window.onPersonalChange`：接收主题、颜色、圆角、语言等个性化变化。

插件页面要使用内核 CSS 变量适配主题，例如 `--user-primary-color`、`--user-bg-1-color`、`--user-readable-surface-color`、`--user-border-color`、`--user-round-enabled`。

## 三、插件后端能力

`plugins/main_demo/impl` 中演示了常用能力：

- 缓存：设置和读取短生命周期数据。
- 持久数据：保存插件自己的轻量键值。
- 数据库：注册模型、自动迁移、CRUD、事务、数据权限查询。
- 事件：订阅普通事件、广播事件、绝对地址事件，推送和取消订阅。
- 语言：注册语言包并按请求语言读取文本。
- 分布式锁：`TryLock` 和 `UnLock`。
- 日志：写入内核日志管线。
- 运行配置：读取插件运行上下文。

建议开发时从 `main_demo` 拷贝一个新插件目录，再逐步替换 `Code`、模块名、页面和动作。

## 四、内核插件表调试数据

主程序插件调试前，需要先在内核数据库的 `plugin` 表中创建同编码插件记录：

| 字段 | 示例 | 说明 |
| --- | --- | --- |
| `Code` | `main_demo` | 必须和目录名、`index.json`、`plugin.Server` 编码一致 |
| `Name` | `主程序插件演示` | 插件显示名 |
| `PluginType` | `1` | 主程序插件 |
| `Status` | `1` | 启用 |
| `Version` | `1.0.1` | 插件版本 |
| `DebugPort` | `7055` | 大于 0 时内核直接代理到该调试端口 |
| `NeedLogin` | `0` 或 `1` | 是否需要登录访问 |
| `IconUrl` | `/page/main_demo/logo.png` | 图标路径 |

可以通过系统内的插件管理页面新增，也可以在数据库中手动插入。编码不一致会导致 `/page/{code}/` 找不到插件。

## 五、配置调试参数

调试主程序插件时，在 `plugins/main_demo/main.go` 中打开或临时加入：

```go
os.Args = []string{"", "7055", "9099", "main_demo", "5173"}
```

参数含义：

| 位置 | 示例 | 含义 |
| --- | --- | --- |
| `os.Args[1]` | `7055` | 插件调试端口，需与内核插件表 `DebugPort` 一致 |
| `os.Args[2]` | `9099` | 内核端口 |
| `os.Args[3]` | `main_demo` | 插件编码 |
| `os.Args[4]` | `5173` | 插件前端调试端口 |

`5173` 为空时，插件会使用 `plugins/main_demo/static` 中的静态文件调试；不为空时，通常代理到插件前端开发服务。

调试顺序：

1. 启动内核，确认端口为 `9099`。
2. 在插件表插入或更新 `main_demo`，设置 `PluginType=1`、`Status=1`、`DebugPort=7055`。
3. 启动插件 Go 进程。
4. 如果使用前端开发服务，进入 `plugins/main_demo/web` 启动 Vite。
5. 访问 `http://127.0.0.1:9099/page/main_demo/`。

## 六、打包结构

主程序插件打包需要准备 `index.json`，可参考 `../../plugins/main_demo/index.json`。字段会被内核读取并写入 `plugin` 表。

基础字段：

```json
{
  "Name": "主程序插件演示",
  "Description": "用于主程序插件demo",
  "CloseDelay": "0",
  "Code": "main_demo",
  "Version": "1.0.1",
  "PluginType": "1",
  "Category": "其他",
  "Interceptors": "",
  "WebHook": "",
  "LimitVersion": "1.0.0",
  "NeedLogin": "0",
  "IconUrl": "/page/main_demo/logo.png",
  "Log": "首发"
}
```

运行目录中主程序插件通常放在：

```text
plugins/main/{code}/
├── index.json
├── logo.png
├── {code}.exe
└── static/
```

Linux/macOS 下可执行文件名通常不带 `.exe`。官方包的目录结构可以通过安装任意官方插件后查看。

## 七、常见问题

- 访问 `/page/main_demo/` 报插件禁用：检查 `plugin.Status` 是否为 `1`。
- 访问后没有走本地调试端口：检查 `plugin.DebugPort` 是否大于 0，并与 `os.Args[1]` 一致。
- 前端 SDK 未注入：确认入口 HTML 引用了 `../../assets/sdk.js`，且页面是从内核 `/page/{code}/` 打开的。
- 插件接口调用超时：检查插件 Go 进程是否启动、端口是否被占用、`Code` 是否一致。

