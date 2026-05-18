# 疑难解答

[返回总入口](starts.md) | [内核业务开发](kernel-business.md) | [架构原理](architecture.md)

## 启动后没有创建数据库表

最常见原因是版本号没有变化。自动迁移只在 `../api/internal/config/version.json` 和运行目录 `init.json` 的 `version` 不一致时执行。

处理：

1. 修改 `../api/internal/config/version.json` 的 `version`，或临时修改运行目录 `init.json` 的 `version`。
2. 重启后端。
3. 确认模型的 `init()` 中调用了 `base.RegisterMigrate(&Model{})`。
4. 确认模型包已被 Go 启动链路导入，否则 `init()` 不会执行。

## 新增字段没有出现在生成页面

检查字段是否满足生成器规则：

- 字段有 `gorm:"column:xxx"`。
- 字段上方有 `// @field name=字段名`。
- 模型文件路径与 `../util/template/code.go` 中的 `menu` 和 `list` 匹配。
- 运行生成器后检查 `web/src/components/common/views/{menu}/{model}` 是否被更新。

## 英文语言包里出现中文

生成器会把中文字段名写入语言包，英文需要后续人工修正。检查：

- `../api/pkg/locales/model-en.json`
- `../web/src/locales/en`

新增前端业务文案时，也要同时补充中文和英文。

## 主程序插件访问不到

检查顺序：

1. `plugin` 表中是否有同编码记录。
2. `Code` 是否与目录名、`index.json`、`plugin.Server(code, ...)` 一致。
3. `PluginType` 是否为 `1`。
4. `Status` 是否为 `1`。
5. 调试时 `DebugPort` 是否大于 0，并与 `os.Args[1]` 一致。
6. 插件 Go 进程是否真的监听了该端口。
7. 访问路径是否为 `/page/{code}/`。

## 主程序插件前端没有走 Vite 调试服务

检查 `os.Args`：

```go
os.Args = []string{"", "7055", "9099", "main_demo", "5173"}
```

`5173` 为空时，插件会使用 `static` 目录；不为空时才会使用插件前端调试端口。

## `$v9os` 不存在或 SDK 不可用

常见原因：

- 页面不是从 V9OS 内核路径打开，而是直接打开本地 HTML。
- SDK 路径错误。主程序插件和前端插件入口通常需要：

```html
<script src="../../assets/sdk.js"></script>
```

- iframe 还没有拿到窗口 ID，部分依赖 `window.__winId` 的能力需要等 SDK 初始化完成。

## 前端插件打不开

检查：

1. 文件是否在运行目录 `plugins/web/{code}`。
2. 是否存在 `index.json` 和 `index.html`。
3. 插件表 `Code` 是否一致。
4. `PluginType` 是否为 `2`。
5. `Status` 是否为 `1`。
6. 访问路径是否为 `/api/webplugin/{code}/`。

## 第三方插件启动失败

检查：

- 当前系统对应脚本是否存在：Windows 需要 `restart.bat` 和 `stop.bat`，Linux/macOS 需要 `restart.sh` 和 `stop.sh`。
- `index.json` 中 `ThirdPort` 是否存在并和服务真实端口一致。
- 插件表 `AccessUrl` 是否配置为可访问地址。
- 服务启动后 30 秒内端口是否可连接。
- `runtime_error` 字段是否记录了启动错误。
- Linux/macOS 下脚本是否能在插件目录内正常执行。

## 远程应用打开空白

远程应用使用 iframe 嵌入。检查远程站点：

- 是否禁止 iframe 嵌入。
- CSP `frame-ancestors` 是否允许 V9OS 域。
- HTTPS 页面是否嵌入了 HTTP 地址。
- 远程系统登录态是否因 Cookie SameSite 或跨域策略失效。

## 插件安装后图标不显示

检查 `IconUrl`：

- 主程序插件可使用 `/page/{code}/logo.png`。
- 前端插件可使用 `/api/webplugin/{code}/logo.png`。
- 应用商店安装时，远程图标可能会被内核快照为 `/api/appstore/img/{code}`。

## 弹窗样式或层级不对

业务页面不要直接使用 `n-modal` 做主流程弹窗。使用 `$wins.addWindow` 让不同外观的窗口行为一致，并让父子窗口层级、关闭、最小化、置顶逻辑由窗口系统处理。

## 分布式下插件没有在当前机器运行

检查：

- 当前是否启用了分布式。
- 插件是否被分配到当前机器白名单。
- 第三方插件的 `FirstMachine` 是否是当前机器 ID。
- 主程序插件是否有本机可用包或调试端口。
- Redis、队列、数据库配置是否符合分布式运行要求。

## 修改配置后没有生效

V9OS 存在本地配置和远程配置。检查运行目录 `init.json`：

- `local: true` 使用本地配置。
- `local: false` 且存在 `remotes` 时使用远程配置。
- `wait_network` 为 true 时，缓存、数据库、队列初始化失败会等待重试。

开发环境建议先使用本地配置，减少远程配置同步造成的判断干扰。

