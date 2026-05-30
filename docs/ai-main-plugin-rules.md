# AI 主程序插件开发规则（PluginType=1）
[返回总入口](starts.md) | [主程序插件开发](plugin-main.md)

> 用途：把本文件和业务需求交给 AI，即可生成 V9OS 主程序插件代码。AI 必须优先遵守本文件。

## 1. 适用范围

主程序插件是 `PluginType = 1` 的插件。它有独立 Go 进程，通过 `github.com/fs185085781/v9os/share/plugin` 调用内核能力，并可附带前端页面，通过 `/page/{code}/` 打开。

以下场景必须优先使用主程序插件：

- 需要 Go 后端进程、业务 API、后台任务、队列、轮询、运行态管理。
- 需要结构化数据库模型、CRUD、事务、数据权限查询。
- 需要缓存、轻量持久数据、分布式锁、事件订阅/广播、日志、语言包。
- 需要对接内核能力、文件能力、网络存储、第三方服务或跨进程能力。
- 前端只是业务界面，真正数据和规则由插件后端承载。

如果业务只有静态页面、报表展示、少量用户偏好键值存储，且不需要 Go 后端，请改用前端插件规则。

## 2. AI 开发主程序插件的固定流程

AI 收到业务需求后必须按此顺序执行：

1. **确定插件编码**：使用小写蛇形命名，例如 `project_manager`。编码必须同时用于目录名、`index.json.Code`、`plugin.Server`、前端 `pluginCode`、页面 URL、图标 URL、表名前缀。
2. **判断插件类型**：只要需要后端模型、事务、后台任务或复杂业务数据，固定使用 `PluginType=1`。
3. **拆业务模块**：按“模型 + 动作 + 页面 + 可选运行器”拆分。例如项目管理可拆成项目、任务、日程、进度、仪表盘。
4. **先设计模型**：检查字段资源限制、SQL 关键字、索引数量、文本数量，确认安全后再写 struct。
5. **写后端骨架**：`go.mod`、`main.go`、`impl/common.go`，并按“每个模型一个文件、每个业务域一个控制器/动作文件”组织代码。
6. **写前端页面**：入口必须引入 SDK，调用 `$v9os.api.pluginPost(code, action, payload)`。
7. **写元数据**：`index.json` 必须完整，`Code` 和 `PluginType` 必须正确。
8. **验证**：`gofmt` 后在插件目录运行 `go test ./...` 或 `go build`。
9. **输出交付说明**：列出文件、动作名、访问路径、调试参数、验证结果。

禁止：复制 demo 后漏改 `main_demo`、硬编码其他插件编码、模型字段超资源、使用不安全列名、前端绕过 `$v9os.api.pluginPost` 调插件后端。

## 3. 标准目录结构

开发目录推荐：

```text
plugins/{code}/
├── go.mod
├── go.sum
├── index.json
├── logo.png 或 logo.svg
├── main.go
├── impl/
│   ├── common.go                 # 常量、通用工具、共享 DTO
│   ├── project_model.go          # 一个模型一个文件
│   ├── project_meta_model.go
│   ├── task_model.go
│   ├── task_meta_model.go
│   ├── member_model.go
│   ├── log_model.go
│   ├── project_action.go         # 一个业务域一个控制器/动作文件
│   ├── task_action.go
│   ├── group_action.go
│   ├── admin_action.go
│   └── 可选：comment_action.go、runner.go、lang.go、event.go、util.go
├── static/
│   ├── index.html
│   ├── logo.png 或 logo.svg
│   ├── css/
│   │   ├── theme.css             # 主题变量和全局样式
│   │   ├── layout.css            # 布局样式
│   │   └── board.css             # 领域样式，按页面/业务拆分
│   ├── js/
│   │   ├── main.js               # 入口，只做初始化和路由挂载
│   │   ├── api.js                # SDK/API 封装
│   │   ├── state.js              # 全局状态
│   │   ├── layout.js             # 布局/导航渲染
│   │   ├── project.js            # 项目领域
│   │   ├── task.js               # 任务领域
│   │   ├── board.js              # 看板视图
│   │   ├── list.js               # 列表视图
│   │   ├── calendar.js           # 日历视图
│   │   ├── gantt.js              # 甘特视图
│   │   ├── modal.js              # 弹窗/抽屉
│   │   └── util.js               # 通用工具
│   └── 其他静态资源或构建产物
└── web/
    └── 可选：Vite/Vue 源码，构建输出到 ../static
```

打包运行目录通常为：

```text
plugins/main/{code}/
├── index.json
├── logo.png
├── {code}.exe     # Windows
└── static/
```

Linux/macOS 可执行文件名通常不带 `.exe`。

## 4. `index.json` 规则

主程序插件必须提供 `index.json`：

```json
{
  "Name": "项目管理",
  "Description": "项目、任务、日程和进度管理插件",
  "CloseDelay": "0",
  "Code": "project_manager",
  "Version": "1.0.0",
  "PluginType": "1",
  "Category": "效率工具",
  "WebHook": "",
  "LimitVersion": "1.0.0",
  "IconUrl": "/page/project_manager/logo.png",
  "Log": "首发"
}
```

强制规则：

- `Code` 必须等于目录名、Go module 建议名、`plugin.Server` 编码和前端 `pluginCode`。
- `PluginType` 必须为字符串 `"1"`。
- `IconUrl` 主程序插件使用 `/page/{code}/logo.png` 或 `/page/{code}/logo.svg`。
- `Version` 从 `1.0.0` 起步；只改规则或说明不应随意提升版本。

## 5. `go.mod` 规则

开发目录内的主程序插件必须引用 `main/share`：

```go
module project_manager

go 1.24.6

require github.com/fs185085781/v9os/share v0.0.0-20251117022245-0f4f6760fbc3

replace github.com/fs185085781/v9os/share => ../../main/share
```

如果需要其他依赖，优先使用标准库；确有必要再新增依赖，并保持插件目录内 `go.sum` 同步。

## 6. `main.go` 入口规则

最小入口：

```go
package main

import (
    "embed"
    _ "project_manager/impl"
    "os"

    "github.com/fs185085781/v9os/share/plugin"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
    // 调试参数：空、插件端口、主程序端口、插件编码、前端调试端口
    os.Args = []string{"", "9210", "9099", "project_manager", ""}
    plugin.Server("project_manager", staticFiles)
}
```

规则：

- `plugin.Server("{code}", staticFiles)` 的编码必须与 `index.json.Code` 一致。
- `impl` 包使用空导入触发 `init()` 注册模型和动作。
- `//go:embed static/*` 要求 `static` 目录至少存在一个文件。
- 如果插件启动后要初始化后台任务，使用：

```go
plugin.ServerAfterAction("project_manager", staticFiles, func() {
    StartTaskRuntime()
})
```

- 如需启动时接收用户/部门/数据权限上下文，使用 `ServerCallDataScope` 或 `ServerCallDataScopeAndAfterAction`。

## 7. 后端代码组织和动作注册规则

### 7.1 强制代码组织

主程序插件后端必须按业务域拆分，禁止把所有模型揉进一个 `model.go`，也禁止把所有动作揉进一个 `action.go`。

强制规则：

- **每个模型一个文件**：例如 `project_model.go`、`task_model.go`、`task_meta_model.go`、`member_model.go`。
- **每个业务域一个控制器/动作文件**：例如 `project_action.go`、`task_action.go`、`group_action.go`、`admin_action.go`、`comment_action.go`。
- 共享常量、DTO、工具函数放在 `common.go`、`dto.go`、`util.go`，不得反向依赖具体 action。
- 单个 action 文件建议不超过 500 行；超过时继续按子域拆分，例如 `task_sort_action.go`、`task_batch_action.go`。
- 模型文件只放模型 struct、模型常量和该模型的注册；不要放业务 CRUD 流程。
- 业务动作文件只放该业务域动作 struct、RunData、注册和小型私有辅助函数；跨域通用逻辑必须上移到 service/util 文件。
- 复杂插件推荐增加 service 层：`project_service.go`、`task_service.go`、`log_service.go`，action 只负责参数解析和返回。

### 7.2 动作实现

每个后端动作必须实现：

```go
type ProjectListAction struct{}

func (a *ProjectListAction) RunData(r *http.Request, param []byte) (any, error) {
    // 业务逻辑
    return result, nil
}
```

注册动作：

```go
func init() {
    plugin.Register("project_list", &ProjectListAction{}, "项目管理", "项目列表")
    plugin.RegisterLogin("project_save", &ProjectSaveAction{}, "项目管理", "保存项目")
}
```

规则：

- 动作名使用小写蛇形或短横风格，项目内保持一致，例如 `project_list`、`task_save`。
- `plugin.Register`：不强制登录，但仍可上报权限元数据。
- `plugin.RegisterLogin`：需要登录；若请求头没有 `userID`，内核返回“请登录后重试”。真实业务默认优先使用 `RegisterLogin`。
- 注册时可传入两个 meta：`feature` 和 `label`，插件启动后会上报到内核权限系统。
- 动作返回 `any`，错误返回 `error`；不要在动作里手写 HTTP JSON 响应。
- 参数读取简单场景用 `plugin.ConvertParam(param)`，复杂结构用 `plugin.Convert(param, &req)`。

参数读取示例：

```go
p := plugin.ConvertParam(param)
id := p.ParamString("ID")
limit := p.ParamInt("Limit")
```

结构体参数示例：

```go
var item ProjectRecord
if err := plugin.Convert(param, &item); err != nil {
    return nil, err
}
```

## 8. 模型注册与字段硬性约束

模型必须嵌入 `plugin.BaseModel`：

```go
type ProjectRecord struct {
    plugin.BaseModel
    ProjectNo   string `gorm:"column:project_no;index"`
    Name        string `gorm:"column:name;index"`
    OwnerName   string `gorm:"column:owner_name;index"`
    Stage       string `gorm:"column:stage;index"`
    RiskLevel   string `gorm:"column:risk_level;index"`
    ProgressPct int    `gorm:"column:progress_pct"`
    StartAt     int64  `gorm:"column:start_at"`
    DueAt       int64  `gorm:"column:due_at"`
    Summary     string `gorm:"column:summary;type:text"`
}

func init() {
    plugin.RegisterModel(&ProjectRecord{}, "project_manager_project")
}
```

### 8.1 必须遵守的模型限制

- 模型字段不允许添加 `json` tag；只保留 `gorm` tag 或必要的数据库 tag。
- 不要使用 `bool` 类型。开关、是否、启用状态统一使用 `int`（如 `0/1`）或字符串枚举。
- 模型字段名和 `gorm:"column:xxx"` 列名都必须规避 SQL 关键字、排序词、聚合函数名、表达式关键字。
- 禁止或强烈不推荐列名：`asc`、`desc`、`and`、`or`、`not`、`in`、`is`、`null`、`like`、`between`、`exists`、`as`、`on`、`by`、`order`、`group`、`having`、`select`、`distinct`、`count`、`sum`、`avg`、`min`、`max`、`case`、`when`、`then`、`else`、`end`、`type`。
- 插件模型最终落到内核插件数据表结构中，字段资源有限：**每个模型常规字段最多 20 个、索引字段最多 5 个、文本字段最多 5 个**。
- 文本字段使用 `gorm:"column:xxx;type:text"`，且文本字段不要再加索引。
- 新增字段前必须评估是否真的需要，避免占用后续扩展空间。

### 8.2 推荐列名写法

- `Type` 改成 `BizType` / `ScheduleType`，列名用 `biz_type` / `schedule_type`。
- `Desc` 改成 `Remark` / `DescriptionText`，列名用 `remark` / `description_text`。
- `Order` 改成 `SortNo`，列名用 `sort_no`。
- `Group` 改成 `GroupName`，列名用 `group_name`。
- `Status` 可用，列名 `status`。
- 时间统一用 `int64` 毫秒/秒时间戳，列名 `start_at`、`due_at`、`report_at`。

## 9. 数据库 CRUD 规则

普通查询：

```go
var items []ProjectRecord
err := plugin.Db("project_manager_project").
    Where("name LIKE ?", "%"+keyword+"%").
    Order("updated_at desc").
    Limit(50).
    Find(&items)
```

带数据权限查询：

```go
err := plugin.DbWithScope("project_manager_project", r).
    Order("updated_at desc").
    Find(&items)
```

规则：

- 登录业务默认用 `DbWithScope(table, r)`；系统级公共数据才用 `Db(table)`。
- 新增 ID 使用 `plugin.Uuid()`，不要依赖数据库自增。
- 更新推荐 `Where("id = ?", id).Updates(map[string]any{...})`。
- 删除推荐先校验 ID，再 `Delete(&Model{})`。
- 列表必须限制数量或分页，避免无上限查询。
- 搜索条件必须参数化，禁止拼接用户输入到 SQL 字符串。

事务：

```go
err := plugin.Transaction(func(tx *plugin.GormDb) error {
    if err := tx.Session("project_manager_project").Create(&project); err != nil {
        return err
    }
    if err := tx.Session("project_manager_task").Create(&task); err != nil {
        return err
    }
    return nil
})
```

规则：

- 事务内切换表必须使用 `tx.Session(table)`。
- 事务内任何错误直接返回，框架会回滚。

## 10. 内核能力使用规则

可用能力：

```go
plugin.SetCache(key, value, minute)
plugin.GetCache(key)
plugin.SetData(key, value)
plugin.GetData(key)
plugin.TryLock(key)
plugin.UnLock(key, token)
plugin.PushEvent(event, payload)
plugin.SubscribeEvent(event, method)
plugin.SubscribeBroadcastEvent(event, method)
plugin.SubscribeAbsoluteURL(event, url)
plugin.UnsubscribeEvent(event, methodOrUrl)
plugin.RegisterLang("zh", map[string]string{"key": "文本"})
plugin.GetText(r, "key")
plugin.Log("info", "message", map[string]any{"plugin": code})
plugin.AppConfig()
```

使用规则：

- 缓存适合短生命周期状态、验证码、任务进度摘要。
- `SetData/GetData` 适合插件级轻量持久键值，不适合复杂业务表。
- 分布式锁必须用 `TryLock` 返回的 token 解锁，解锁调用 `UnLock`。
- 事件回调动作也必须注册成插件动作。
- 日志要带结构化字段，如插件编码、动作名、业务 ID。
- 多语言文案放入 `RegisterLang`，请求内取 `plugin.GetText(r, key)`。

## 11. 前端页面接入规则

主程序插件前端运行在 `/page/{code}/`。入口 HTML 必须引入 SDK：

```html
<script src="../../assets/sdk.js"></script>
```

调用插件后端：

```js
const pluginCode = "project_manager";
const data = await $v9os.api.pluginPost(pluginCode, "project_list", { keyword: "" }, "err");
```

规则：

- 主程序插件前端调用后端只能用 `$v9os.api.pluginPost(code, action, payload, showType)`。
- `showType` 可用：`""` 不提示、`"err"` 错误提示、`"ok"` 成功提示、`"okerr"` 成功/失败都提示、`"json"` 返回完整响应。
- 页面必须从内核 `/page/{code}/` 打开；本地 file 直接打开时 `$v9os` 不可用。
- 子窗口使用内核窗口系统：

```js
$v9os.invoke("$wins", "addWindow", {
  width: 860,
  height: 680,
  title: "编辑项目",
  iframeUrl: `${$v9os.host}/page/project_manager/?page=project-edit&id=${id}`
}, window.__winId);
```

- 父子窗口通信使用 `$v9os.event.on` / `$v9os.event.emit`，组件销毁或页面关闭时使用 `$v9os.event.off`。
- 右键菜单使用 `$v9os.contextMenu.show()` 和 `$v9os.contextMenu.onAction()`。
- 消息框使用 `$v9os.msg.alert`、`$v9os.msg.confirm`、`$v9os.msg.prompt` 或 `$v9os.msg.success/error`。
- 需要文件选择时可使用 `$v9os.file.selectFile`、`$v9os.file.selectLongDir`、`$v9os.file.saveFile`。

## 12. 主题和 UI 规则

前端必须适配 V9OS 主题变量：

```css
:root {
  --app-primary: var(--user-primary-color, #2080f0);
  --app-primary-text: var(--user-primary-text-color, #fff);
  --app-bg: var(--user-bg-1-color, #f6f8fb);
  --app-surface: var(--user-readable-surface-color, #fff);
  --app-border: var(--user-border-color, rgba(0,0,0,.12));
  --app-text: var(--user-text-1-color, #172033);
  --app-muted: var(--user-text-2-color, #667085);
  --app-radius: calc(var(--user-round-enabled, 1) * 8px);
}
```

监听个性化变化：

```js
window.onPersonalChange = (settings, theme) => {
  document.documentElement.dataset.theme = settings.Theme === "dark" ? "dark" : "light";
};
```

如果使用 Naive UI，应将 SDK 注入的 CSS 变量映射到 `themeOverrides`。

## 13. Vite/Vue 前端规则（可选）

如果使用 `web/` 源码和 Vite：

```js
export default defineConfig(({ command }) => ({
  base: command === "serve" ? "/page/project_manager/" : "./",
  build: { outDir: "../static" },
  server: {
    host: "0.0.0.0",
    port: 5210,
    hmr: { host: "localhost", port: 5210, protocol: "ws" }
  }
}));
```

规则：

- `base` 开发时必须是 `/page/{code}/`，构建时用 `./`。
- 构建产物输出到插件根目录 `static/`。
- `main.go` 调试参数第五位填写 Vite 端口，例如 `5210`，空字符串则使用 `static/`。

## 14. 调试规则

主程序插件调试前，内核插件表必须有同编码记录：

| 字段 | 示例 | 说明 |
| --- | --- | --- |
| `Code` | `project_manager` | 必须一致 |
| `Name` | `项目管理` | 显示名 |
| `PluginType` | `1` | 主程序插件 |
| `Status` | `1` | 启用 |
| `Version` | `1.0.0` | 版本 |
| `DebugPort` | `9210` | 大于 0 时内核代理到该插件端口 |
| `IconUrl` | `/page/project_manager/logo.png` | 图标 |

调试顺序：

1. 启动内核，确认端口 `9099`。
2. 插件表插入/更新 `Code={code}`、`PluginType=1`、`Status=1`、`DebugPort={pluginPort}`。
3. 启动插件 Go 进程。
4. 如使用 Vite，启动 `web` 前端开发服务。
5. 访问 `http://127.0.0.1:9099/page/{code}/`。

## 15. 后端动作命名建议

常见业务动作：

```text
dashboard
project_list / project_detail / project_save / project_delete
task_list / task_detail / task_save / task_update_status / task_delete
schedule_list / schedule_save / schedule_delete
progress_list / progress_save / progress_summary
setting_get / setting_save
```

规则：

- 列表动作用 `*_list`，详情用 `*_detail`，保存用 `*_save`，删除用 `*_delete`。
- 修改单个字段可以用语义动作，例如 `task_update_status`。
- 后台任务可用 `job_start`、`job_pause`、`job_status`。

## 16. 错误处理和返回规则

- 参数缺失必须返回明确错误，例如 `return nil, errors.New("项目 ID 不能为空")`。
- 可恢复业务失败可返回 `{Ok:false, Message:"..."}`，但真正错误应返回 `error`。
- 不要 panic。
- 不要吞掉数据库错误。
- 删除和更新前必须校验 ID。
- 批量操作必须限制数量。

## 17. 安全规则

- 所有用户输入进入查询都必须使用 `?` 参数化。
- 需要登录和数据隔离的动作使用 `RegisterLogin` + `DbWithScope`。
- 不把 token、密码、密钥写入前端或日志。
- 文件/路径/URL 参数必须做空值、协议、范围校验。
- 后台任务启动前建议用分布式锁避免重复运行。

## 18. 最小后端模板

```go
package impl

import (
    "errors"
    "net/http"

    "github.com/fs185085781/v9os/share/plugin"
)

const (
    pluginCode = "project_manager"
    tableProject = "project_manager_project"
)

type ProjectRecord struct {
    plugin.BaseModel
    Name      string `gorm:"column:name;index"`
    OwnerName string `gorm:"column:owner_name;index"`
    Status    string `gorm:"column:status;index"`
    Summary   string `gorm:"column:summary;type:text"`
}

type ProjectListAction struct{}

func (a *ProjectListAction) RunData(r *http.Request, param []byte) (any, error) {
    p := plugin.ConvertParam(param)
    keyword := p.ParamString("keyword")
    db := plugin.DbWithScope(tableProject, r).Order("updated_at desc").Limit(100)
    if keyword != "" {
        db = db.Where("name LIKE ?", "%"+keyword+"%")
    }
    var items []ProjectRecord
    if err := db.Find(&items); err != nil {
        return nil, err
    }
    return items, nil
}

type ProjectSaveAction struct{}

func (a *ProjectSaveAction) RunData(r *http.Request, param []byte) (any, error) {
    var item ProjectRecord
    if err := plugin.Convert(param, &item); err != nil {
        return nil, err
    }
    if item.Name == "" {
        return nil, errors.New("项目名称不能为空")
    }
    if item.ID == "" {
        item.ID = plugin.Uuid()
        if err := plugin.DbWithScope(tableProject, r).Create(&item); err != nil {
            return nil, err
        }
        return item, nil
    }
    if err := plugin.DbWithScope(tableProject, r).Where("id = ?", item.ID).Select("*").Updates(&item); err != nil {
        return nil, err
    }
    return item, nil
}

func init() {
    plugin.RegisterModel(&ProjectRecord{}, tableProject)
    plugin.RegisterLogin("project_list", &ProjectListAction{}, "项目管理", "项目列表")
    plugin.RegisterLogin("project_save", &ProjectSaveAction{}, "项目管理", "保存项目")
}
```

## 19. 最小前端调用模板

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <script src="../../assets/sdk.js"></script>
  <title>项目管理</title>
</head>
<body>
  <button id="refresh">刷新项目</button>
  <pre id="output"></pre>
  <script>
    const pluginCode = "project_manager";
    async function api(action, payload = {}, show = "err") {
      return await $v9os.api.pluginPost(pluginCode, action, payload, show);
    }
    async function refresh() {
      const data = await api("project_list", { keyword: "" }, "err");
      document.querySelector("#output").textContent = JSON.stringify(data || [], null, 2);
    }
    window.onPersonalChange = (settings) => {
      document.documentElement.dataset.theme = settings.Theme === "dark" ? "dark" : "light";
    };
    document.querySelector("#refresh").addEventListener("click", refresh);
    refresh();
  </script>
</body>
</html>
```

## 20. 生成后验收清单

AI 完成主程序插件后必须自检：

- [ ] `index.json.Code`、目录名、`plugin.Server`、前端 `pluginCode` 完全一致。
- [ ] `index.json.PluginType` 为 `"1"`。
- [ ] `main.go` 有 `//go:embed static/*`，且 `static` 目录不为空。
- [ ] 所有模型已 `plugin.RegisterModel`。
- [ ] 模型无 `json` tag，无 `bool` 字段。
- [ ] 每个模型常规字段 ≤ 20、索引字段 ≤ 5、文本字段 ≤ 5。
- [ ] 列名避开 SQL 关键字和 `type/desc/order/group/end` 等高风险词。
- [ ] 登录业务使用 `RegisterLogin` 和 `DbWithScope`。
- [ ] 查询使用参数化，列表有限制或分页。
- [ ] 前端入口引入 `../../assets/sdk.js`。
- [ ] 前端按领域拆分文件，入口、API、状态、布局、项目、任务、看板、列表、日历、甘特、弹窗、工具各自独立；禁止把全部逻辑堆进一个 `app.js`。
- [ ] 前端使用 `$v9os.api.pluginPost` 调后端。
- [ ] 前端使用 V9OS CSS 变量适配主题。
- [ ] 已运行 `gofmt`。
- [ ] 已运行 `go test ./...` 或 `go build` 并记录结果。
