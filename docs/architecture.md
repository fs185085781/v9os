# 架构原理

[返回总入口](starts.md) | [内核业务开发](kernel-business.md) | [疑难解答](troubleshooting.md)

本页概述 V9OS 开发时最容易影响判断的几条链路：启动、配置、迁移、插件、窗口 SDK 和分布式运行。

## 一、启动链路

后端入口在：

- `../api/cmd/console/main.go`
- `../api/cmd/gui/main.go`

启动时会创建配置、日志、缓存、数据库、队列、分布式 Provider、插件管理器和 Gin 路由。核心初始化在 `../api/internal/server/server.go`。

大致顺序：

1. 读取 `version.json` 和运行目录 `init.json`，得到机器级配置。
2. 初始化缓存，默认支持本地文件缓存，也可切换 Redis。
3. 初始化数据库，默认 SQLite，也支持 MySQL、PostgreSQL、SqlServer、GaussDB、ClickHouse 等。
4. 如果版本变化，执行自动迁移。
5. 初始化消息队列，默认内存队列，也可使用 Redis 或 RocketMQ。
6. 初始化分布式运行时。
7. 初始化插件管理器。
8. 注册 API、插件代理、静态资源和前端入口。

## 二、配置与版本

`../api/internal/config/version.json` 会被 Go `embed` 进程序。运行目录的 `init.json` 是机器级配置，包含端口、机器 ID、远程配置地址、是否本地配置、版本等。

版本比较规则：

- `init.json.version != version.json.version`：认为需要更新，设置 `NeedUpdate`。
- `NeedUpdate = true` 后，启动中执行数据库迁移。
- 启动完成后的延迟函数会把运行目录 `init.json.version` 写回当前内核版本。

因此，开发新增模型或字段后，如果没有变化版本号，数据库可能不会自动创建新表或新列。

## 三、数据库自动迁移

迁移入口是 `../api/internal/database/auto_migrate.go`。

普通内核模型：

1. 模型包 `init()` 调用 `base.RegisterMigrate(&Model{})`。
2. 注册对象进入 IOC 中的模型集合。
3. `database.AutoMigrate()` 从集合取出模型并执行 GORM `AutoMigrate`。

插件数据模型：

1. 通过 `base.RegisterPluginData` 注册虚拟插件数据结构和物理表。
2. 迁移时按指定表名执行 `AutoMigrate`。

迁移完成后会检查是否存在 ID 为 1 的管理员，不存在则创建默认管理员 `admin / 123456`。

## 四、代码生成器

生成器在 `../util/template/code.go`，它解析 Go AST 和结构体注释，不依赖运行中的后端。

它识别：

- `@model name=...`
- `@field name=...`
- `@select key=value`
- `@datetime`
- `@textarea`
- `gorm:"column:..."`

生成目标包括后端控制器、前端 CRUD 页面和模型语言包。它适合快速生成标准业务表的基础功能，复杂业务仍需要手工补充。

## 五、插件类型与访问路径

插件元数据由 `../api/internal/model/plugin/plugin.go` 定义，核心字段包括：

- `Code`：插件编码。
- `PluginType`：`1` 主程序插件、`2` 前端插件、`3` 第三方插件、`4` 远程应用。
- `Status`：启用状态。
- `Version`：插件版本。
- `AccessUrl`：第三方插件或远程应用地址。
- `DebugPort`：主程序插件调试端口。

访问路径：

| 类型 | 管理器 | 路径 |
| --- | --- | --- |
| 主程序插件 | `main_manage.go` | `/page/{code}/` |
| 前端插件 | `web_manage.go` | `/api/webplugin/{code}/` |
| 第三方插件 | `third_manage.go` | `/api/thirdplugin/{code}` |
| 远程应用 | `frame_manage.go` | `AccessUrl` |

主程序插件如果设置了 `DebugPort`，内核直接把请求代理到调试端口；否则检查运行目录插件可执行文件，不存在时尝试安装对应版本包。

## 六、插件包与 index.json

插件安装时会读取包根目录 `index.json`，转换为插件表记录。通用字段包括：

- `Name`
- `Description`
- `CloseDelay`
- `Code`
- `Status`
- `Version`
- `PluginType`
- `WebHook`
- `LimitVersion`
- `IconUrl`
- `DebugPort`
- `ThirdPort`

第三方插件额外依赖 `ThirdPort` 和启动/停止脚本。主程序插件额外依赖可执行文件。前端插件额外依赖 `index.html`。

## 七、前端窗口与 SDK

前端窗口状态在 `../web/src/stores/windows.js` 中维护。`$wins.addWindow` 会创建或复用窗口，设置父子窗口关系、位置、层级和 iframe 地址。

插件 iframe 通过 `initPostMessage` 和 SDK 通信：

- iframe 请求当前窗口 ID。
- iframe 通过 `iframe-invoke` 调用主窗口对象方法。
- iframe 通过 `iframe-event-on` 订阅主窗口事件。
- 主窗口把个性化设置通过 `personalChange` 发送给插件。

`../web/public/assets/sdk.js` 会注入 `$v9os`，并负责主题变量、消息提示、插件请求、事件、右键菜单、剪贴板等桥接能力。

## 八、前端主题与多外观

V9OS 前端支持后台、Win10、macOS、Deepin、Pad 等外观。共享业务页面位于 `../web/src/components/common`，各外观窗口和桌面壳位于对应目录。

开发约束：

- 不要把颜色写死成单一主题，优先使用用户 CSS 变量。
- 圆角使用 `--user-round-enabled` 控制。
- 深色模式要检查文字、边框、表格、弹窗和悬浮层。
- 弹窗使用 `$wins.addWindow`，避免直接使用 `n-modal` 形成外观不一致。
- 新增文案要进入多语言文件。

## 九、分布式运行

V9OS 的设计目标是同一套代码可运行在单体或分布式环境。相关点：

- 缓存可使用本地或 Redis。
- 队列可使用内存、Redis 或 RocketMQ。
- 数据库需要确认所选驱动是否支持目标分布式能力。
- 主程序插件和第三方插件有机器、白名单、调试端口、运行地址等约束。
- 第三方插件在分布式下通常只由指定机器运行。

开发共享模块时，不要假设所有能力都在同一进程、同一机器或同一内存中。

