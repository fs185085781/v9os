# V9OS 开发使用教程

本文档是 V9OS 二次开发和插件开发的总入口。建议先按 `../README.md` 的快速开始启动一遍项目，再根据开发目标进入对应章节。

## 文档导航

| 目标 | 文档 |
| --- | --- |
| 新增内核业务功能、使用代码生成器、触发数据库迁移 | [内核业务开发](kernel-business.md) |
| 开发 Go 后端能力 + 插件前端的主程序插件 | [主程序插件开发](plugin-main.md) |
| 开发仅前端资源的插件 | [前端插件开发](plugin-web.md) |
| 把任意可运行 Web 程序包装为插件 | [第三方插件开发](plugin-third.md) |
| 接入外部站点或已有系统 | [远程应用](remote-app.md) |
| 理解启动、迁移、插件、前端窗口和分布式设计 | [架构原理](architecture.md) |
| 排查数据库表未创建、插件打不开、SDK 不可用等问题 | [疑难解答](troubleshooting.md) |

## 推荐阅读路线

1. 内核功能开发者：先读 [内核业务开发](kernel-business.md)，再读 [架构原理](architecture.md) 中的“启动与迁移链路”和“前端窗口与 SDK”。
2. 插件开发者：先根据插件类型阅读 [主程序插件开发](plugin-main.md)、[前端插件开发](plugin-web.md) 或 [第三方插件开发](plugin-third.md)，再读 [架构原理](architecture.md) 中的“插件类型与访问路径”。
3. 运维和部署人员：先读 [架构原理](architecture.md)，再读 [疑难解答](troubleshooting.md)。

## 开发入口速览

### 内核业务

内核业务适合放在主程序中长期维护的系统能力。标准流程是：

1. 在 `../api/internal/model` 下创建模型，并在 `init()` 中调用 `base.RegisterMigrate(&YourModel{})`。
2. 修改 `../util/template/code.go`，把要生成的模型名和菜单包名加入生成器。
3. 在 VSCode 运行 `V9os Code`，或进入 `../util` 后执行等价的生成命令。
4. 修改 `../api/internal/config/version.json` 的版本号，或修改运行目录 `init.json` 中的版本号，让两者产生差异。
5. 在 VSCode 启动 `Debug V9os`，或进入 `../api` 后执行等价的后端启动命令。
6. 检查数据库表、后端接口、前端页面和语言包是否生成并可用。

完整说明见 [内核业务开发](kernel-business.md)。

### 插件类型

V9OS 当前主要有四类扩展：

| 类型 | `PluginType` | 适用场景 | 访问路径 |
| --- | --- | --- | --- |
| 主程序插件 | `1` | 需要 Go 后端能力、缓存、队列、数据库、事件等内核能力 | `/page/{code}/` |
| 前端插件 | `2` | 只有 HTML/CSS/JS 或任意前端构建产物 | `/api/webplugin/{code}/` |
| 第三方插件 | `3` | 独立进程 Web 程序，需要启动/停止脚本 | `/api/thirdplugin/{code}/` |
| 远程应用 | `4` | 外部 URL 或云端系统，以 iframe 方式接入 | 远程地址 |

插件记录保存在内核数据库的 `plugin` 表。调试或手动安装时，插件编码 `Code` 必须和目录名、`index.json` 中的 `Code` 保持一致。

## 通用约定

- 文档中的路径均以 `main` 目录为参照。
- 代码和文档请使用 UTF-8 保存。
- 前端插件和主程序插件前端都需要引入 SDK：`<script src="../../assets/sdk.js"></script>`。如果目录层级不同，请按实际入口相对到内核 `assets/sdk.js`。
- 修改 `api/internal/inface` 和 `web/src/components/common/inface` 时，要考虑社区版和企业版插件隔离，避免社区版删除或裁剪时大面积引用报错。
- 修改 `web` 时，要考虑主题色、圆角、深浅色、透明度和多语言；弹窗优先通过 `$wins.addWindow` 打开。
