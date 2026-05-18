package local_billing

import "context"

// Status 描述“用户付费授权”模块当前是否真正可用。
// 这个模块不是有 EE 代码就强制开启，必须同时满足 user_ee 存在、
// 管理员已开启配置、商业授权包含 local_billing.enable，最终 Enabled 才会为 true。
type Status struct {
	// AdminEnabled 表示管理员是否在本机配置中开启用户付费授权。
	AdminEnabled bool `json:"adminEnabled"`
	// OfficialAuthorized 表示商业授权中是否允许启用 local_billing.enable。
	OfficialAuthorized bool `json:"officialAuthorized"`
	// Enabled 是前面条件汇总后的最终开关，业务代码只需要看这个值。
	Enabled bool `json:"enabled"`
}

// Provider 是 local_billing 对外暴露的最小抽象。
// 社区版默认实现只返回不可用状态，EE 目录存在时由 fuse 替换成真实实现。
type Provider interface {
	// Status 返回当前模块的拆分状态，便于前端展示“缺 EE / 缺 user_ee / 未授权 / 未开启”。
	Status(ctx context.Context) Status
}
