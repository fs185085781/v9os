# 内核业务开发

[返回总入口](starts.md) | [架构原理](architecture.md) | [疑难解答](troubleshooting.md)

内核业务适合放入主程序代码仓库中，由 V9OS 内核统一编译、迁移、鉴权和发布。典型功能包括系统管理、用户相关功能、插件管理、审计日志和长期稳定的业务模块。

## 一、创建数据库模型

模型放在 `../api/internal/model` 下，建议按业务域分包，例如：

```text
api/internal/model/
├── system/
├── user/
└── your_menu/
    └── your_model.go
```

一个可被生成器识别和自动迁移的模型需要包含：

- `base.BaseModel`：提供 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`。
- `gorm:"column:..."`：生成器只处理带 `column` 的字段。
- `// @model name=...`：模型中文名。
- `// @field name=...`：字段中文名。
- 可选标记：`// @select 1=启用 0=禁用`、`// @datetime`、`// @textarea`。
- `TableName()`：明确表名。
- `init()` 中调用 `base.RegisterMigrate(&YourModel{})`。

示例：

```go
package your_menu

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=业务任务
type Task struct {
    base.BaseModel
    // @field name=任务名称
    Name string `gorm:"column:name;size:100"`
    // @field name=状态
    // @select 1=启用 0=禁用
    Status int `gorm:"column:status" json:"Status,string"`
    // @field name=备注
    // @textarea
    Remark string `gorm:"column:remark;type:text"`
}

func (t *Task) TableName() string {
    return "task"
}

func init() {
    base.RegisterMigrate(&Task{})
}
```

注意：新增模型文件需要被 Go 编译器引用到，才能执行 `init()`。放在已有会被导入的模型包中最稳妥；新增包时需要确认启动链路或聚合导入中已经引入该包。

## 二、配置代码生成器

生成器入口在 `../util/template/code.go`。当前逻辑会读取：

```go
menu := "user"
list = []string{
    "desktop_app",
}
```

开发新功能时，把 `menu` 改为模型所在包名，把 `list` 改为模型文件名去掉 `.go` 后的蛇形名称。例如模型文件为：

```text
api/internal/model/your_menu/task.go
```

则配置为：

```go
menu := "your_menu"
list = []string{
    "task",
}
```

生成器会读取模型并生成或更新：

- `../api/pkg/locales/model-zh.json`
- `../api/pkg/locales/model-en.json`
- `../api/internal/controller/api/{menu}/{model}.go`
- `../web/src/components/common/views/{menu}/{model}/index.vue`
- `../web/src/components/common/views/{menu}/{model}/edit.vue`
- `../web/src/components/common/views/{menu}/{model}/detail.vue`
- `../web/src/components/common/views/{menu}/{model}/import.vue`

## 三、运行生成器

优先使用 VSCode 中的 `V9os Code` 调试配置。如果当前工作区没有该配置，可以进入 `../util` 执行等价命令：

```bash
go build -o v9os-code.exe template/code.go
./v9os-code.exe
```

Linux/macOS 下可改为：

```bash
go build -o v9os-code template/code.go
./v9os-code
```

不要直接使用 `go run template/code.go` 作为替代方式。生成器会通过可执行文件所在目录反推 `main` 目录，`go run` 的临时目录可能导致路径定位错误。

生成后需要检查：

- 控制器是否出现在 `api/internal/controller/api/{menu}`。
- 前端页面是否出现在 `web/src/components/common/views/{menu}/{model}`。
- 中英文语言包是否新增了 `model.{model}` 节点。
- 生成器是否覆盖了你手工改过的同名页面；生成前建议确认工作区状态。

## 四、触发数据库迁移

迁移由版本号变化触发。内核启动时会读取嵌入的 `../api/internal/config/version.json`，并与运行目录 `init.json` 中的 `version` 比较：

- 两者不同：设置 `NeedUpdate`，启动后执行 `database.AutoMigrate()`，再把运行目录 `init.json` 更新为当前版本。
- 两者相同：认为无需迁移，不会自动建新表。

开发时可选择任一种方式让版本号产生差异：

1. 修改 `../api/internal/config/version.json` 的 `version`，例如从 `1.0.000` 改为 `1.0.001-dev`。
2. 修改运行目录中的 `init.json` 的 `version`，例如临时改为 `0`。

只要版本号有变化即可，不要求语义版本递增。团队协作时建议统一修改 `version.json`，避免每个人本地行为不一致。

## 五、启动和验证

优先使用 VSCode 中的 `Debug V9os`。如果没有该调试配置，可以进入 `../api` 执行：

```bash
go run cmd/console/main.go
```

验证清单：

- 后端启动日志中没有数据库、缓存、队列初始化错误。
- 数据库中出现新表或新增字段。
- `plugin`、`user`、`system` 等基础表存在，首次启动会创建默认管理员 `admin / 123456`。
- 访问前端开发端口后，菜单或页面能够打开生成的页面。
- 生成的分页、保存、详情、删除、导入、导出接口可调用。

## 六、继续定制业务

生成器给的是标准 CRUD 起点。常见后续工作包括：

- 在生成的控制器中补充业务校验、事务、权限和审计字段。
- 在前端 `index.vue` 调整搜索项、表格列、批量动作和窗口尺寸。
- 在 `edit.vue` 和 `detail.vue` 中替换控件，例如选择器、时间选择器、富文本或文件上传。
- 补充 `model-en.json` 的英文翻译，生成器对英文缺失项会先使用中文字段名。
- 修改菜单注册或桌面快捷方式，让入口出现在目标外观中。

## 七、多语言和分布式注意事项

- 后端返回给用户看的文本优先走语言包，不要只写固定中文。
- 前端新增文案要进入 `web/src/locales`，页面中用 `$t` 或项目已有 i18n 工具读取。
- 涉及缓存、队列、锁、插件运行状态时，要考虑本地模式和 Redis/RocketMQ 等分布式模式。
- 涉及机器、插件白名单、远程节点的逻辑时，要确认单机和分布式部署下都能运行。
