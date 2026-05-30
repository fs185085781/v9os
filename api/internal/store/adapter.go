package store

type Store interface {
	PluginDownloadUrl(code, version, os, arch string) string
	KernelDownloadUrl(version, os, arch, buildType, releaseType string) string
	CheckKernelUpdate(req KernelCheckRequest) (*KernelCheckResponse, error)
	GetCategories() ([]Category, error)
	GetAppsByCategory(category string, page, pageSize int) (*AppListResult, error)
	SearchApps(keyword string, page, pageSize int) (*AppListResult, error)
	GetAppDetail(code string) (*AppInfo, error)
	GetAppInstallDetail(code string) (*AppInstallInfo, error)
	GetAppVersions(code string) ([]AppVersion, error)
	GetFontUrl(font string, ui string) string
	GetFontOptions() ([]FontOption, error)
}

type KernelCheckRequest struct {
	Version     string `json:"version"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	BuildType   string `json:"buildType"`
	ReleaseType string `json:"releaseType"`
}

type KernelCheckResponse struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseType    string `json:"releaseType"`
	Changelog      string `json:"changelog"`
	DownloadURL    string `json:"downloadUrl"`
	PackageName    string `json:"packageName"`
	PackageSize    int64  `json:"packageSize"`
	PackageHash    string `json:"packageHash"`
	Message        string `json:"message"`
}

type AppPackage struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	PackageName string `json:"packageName"`
	PackagePath string `json:"packagePath"`
	PackageHash string `json:"packageHash"`
	PackageSize int    `json:"packageSize"`
	DownloadURL string `json:"downloadUrl"`
}

type Category struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}
type AppInfo struct {
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	IconUrl      string   `json:"iconUrl"`
	Author       string   `json:"author"`
	Version      string   `json:"version"`
	StoreVersion string   `json:"storeVersion"`
	Category     string   `json:"category"`
	Screenshots  []string `json:"screenshots"`
	PluginType   int      `json:"pluginType"`
	LimitVersion string   `json:"limitVersion"`

	Packages         []AppPackage `json:"packages"`
	Installable      bool         `json:"installable"`
	InstallReason    string       `json:"installReason"`
	Installed        bool         `json:"installed"`
	InstalledVersion string       `json:"installedVersion"`
}

type AppInstallInfo struct {
	AppInfo
	AccessUrl string `json:"accessUrl"`
}

type AppVersion struct {
	Version   string       `json:"version"`
	Changelog string       `json:"changelog"`
	CreatedAt string       `json:"createdAt"`
	Packages  []AppPackage `json:"packages"`
}

type AppListResult struct {
	Data  []AppInfo `json:"data"`
	Total int64     `json:"total"`
}

type FontOption struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}
