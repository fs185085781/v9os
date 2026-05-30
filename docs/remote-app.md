# 远程应用

[返回总入口](starts.md) | [第三方插件开发](plugin-third.md) | [架构原理](architecture.md)

远程应用是 `PluginType = 4` 的插件，用于把已有 URL 接入 V9OS 应用商店和窗口系统。它不需要本地插件目录，也不需要启动脚本。

## 一、适用场景

- 已部署的 SaaS、BI、监控面板、文档系统。
- 公司内网系统入口。
- 不需要 V9OS 托管进程生命周期的 Web 应用。
- 快速把外部页面加入桌面或后台应用入口。

## 二、添加方式

直接在应用商店添加远程应用，填写访问地址即可。内核会保存一条插件数据，类型为远程 iframe。

关键字段：

| 字段 | 说明 |
| --- | --- |
| `Name` | 应用名称 |
| `PluginType` | `4` |
| `AccessUrl` | 远程应用完整 URL |
| `Status` | `1` 为启用 |
| `IconUrl` | 应用图标 |

远程应用的插件编码可由访问地址生成，也可以由安装流程保存。打开时前端会直接把 `AccessUrl` 作为 iframe 地址。

## 三、远程页面要求

远程应用本质是 iframe 嵌入，需要远程站点允许被 V9OS 页面加载：

- 远程服务不能设置拒绝嵌入的 `X-Frame-Options: DENY` 或不匹配的 `SAMEORIGIN`。
- 如果使用 CSP，需要允许 V9OS 所在域通过 `frame-ancestors` 嵌入。
- HTTPS 页面中嵌入 HTTP 远程地址可能被浏览器拦截。
- 远程登录态、跨域 Cookie、SameSite 策略由远程系统自行处理。

## 四、选择建议

- 只接 URL：用远程应用。
- 要随 V9OS 启停本地服务：用 [第三方插件](plugin-third.md)。
- 要调用内核 Go SDK 能力：用 [主程序插件](plugin-main.md)。
- 只有前端静态资源并想使用 V9OS SDK：用 [前端插件](plugin-web.md)。

