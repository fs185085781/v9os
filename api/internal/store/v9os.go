package store

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/fs185085781/v9os/internal/logger"
)

type v9osStore struct {
	host   string
	client *http.Client
}

func (v *v9osStore) GetFontUrl(font string, ui string) string {
	return v.host + "/api/font/" + url.PathEscape(font) + "/" + url.PathEscape(ui)
}

func (v *v9osStore) GetFontOptions() ([]FontOption, error) {
	var result []FontOption
	err := v.getJSON("/api/font/select", nil, &result)
	return result, err
}

func (v *v9osStore) PluginDownloadUrl(code string, version string, os string, arch string) string {
	return v.host + "/api/store/download/" + code + "/" + version + "/" + os + "/" + arch
}

func (v *v9osStore) KernelDownloadUrl(version, os, arch, buildType, releaseType string) string {
	return v.host + "/api/kernel/download/" + url.PathEscape(version) + "/" + url.PathEscape(os) + "/" + url.PathEscape(arch) + "/" + url.PathEscape(buildType) + "/" + url.PathEscape(releaseType)
}

func (v *v9osStore) CheckKernelUpdate(req KernelCheckRequest) (*KernelCheckResponse, error) {
	params := url.Values{}
	params.Set("version", req.Version)
	params.Set("os", req.OS)
	params.Set("arch", req.Arch)
	params.Set("build_type", req.BuildType)
	params.Set("release_type", req.ReleaseType)
	var result KernelCheckResponse
	err := v.getJSON("/api/kernel/check", params, &result)
	return &result, err
}

func (v *v9osStore) GetCategories() ([]Category, error) {
	var result []Category
	err := v.getJSON("/api/store/categories", nil, &result)
	return result, err
}

func (v *v9osStore) GetAppsByCategory(category string, page, pageSize int) (*AppListResult, error) {
	params := url.Values{}
	params.Set("category", category)
	params.Set("page", strconv.Itoa(page))
	params.Set("pageSize", strconv.Itoa(pageSize))
	var result AppListResult
	err := v.getJSON("/api/store/apps", params, &result)
	return &result, err
}

func (v *v9osStore) SearchApps(keyword string, page, pageSize int) (*AppListResult, error) {
	params := url.Values{}
	params.Set("keyword", keyword)
	params.Set("page", strconv.Itoa(page))
	params.Set("pageSize", strconv.Itoa(pageSize))
	var result AppListResult
	err := v.getJSON("/api/store/search", params, &result)
	return &result, err
}

func (v *v9osStore) GetAppDetail(code string) (*AppInfo, error) {
	var result AppInfo
	err := v.getJSON("/api/store/app/"+code, nil, &result)
	return &result, err
}

func (v *v9osStore) GetAppVersions(code string) ([]AppVersion, error) {
	var result []AppVersion
	err := v.getJSON("/api/store/app/"+code+"/versions", nil, &result)
	return result, err
}

// getJSON 通用GET请求,解析远程API的JSON响应
func (v *v9osStore) getJSON(path string, params url.Values, dest interface{}) error {
	u := v.host + path
	if params != nil && len(params) > 0 {
		if strings.Contains(u, "?") {
			u += "&" + params.Encode()
		} else {
			u += "?" + params.Encode()
		}
	}
	resp, err := v.client.Get(u)
	if err != nil {
		return fmt.Errorf("请求应用商店失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("应用商店返回错误状态: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取应用商店响应失败: %w", err)
	}
	// 远程API返回格式: {"code":0,"data":...}
	var apiResp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 尝试直接解析为目标类型
		return json.Unmarshal(body, dest)
	}
	if apiResp.Code != 0 {
		return fmt.Errorf("应用商店错误: %s", apiResp.Msg)
	}
	return json.Unmarshal(apiResp.Data, dest)
}

func newV9osStore(host string, tmpLog logger.Logger) (Store, error) {
	res := &v9osStore{
		host:   host,
		client: &http.Client{},
	}
	tmpLog.Println("[v9os数据源]已初始化")
	return res, nil
}
