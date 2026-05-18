package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"sort"
	"time"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/inface/official_license"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/gin-gonic/gin"
)

type LicenseController struct {
	*controller.BaseController
}

type officialLicenseFeatureView struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	AuthType  string `json:"authType"`
	Quantity  int64  `json:"quantity"`
	ExpiredAt int64  `json:"expiredAt"`
	Status    string `json:"status"`
	HasCipher bool   `json:"hasCipher"`
}

type officialLicenseSnapshotView struct {
	Authorized      bool                         `json:"authorized"`
	AuthID          string                       `json:"authId"`
	StartAt         int64                        `json:"startAt"`
	EndAt           int64                        `json:"endAt"`
	FeatureCount    int                          `json:"featureCount"`
	Features        []officialLicenseFeatureView `json:"features"`
	UnavailableText string                       `json:"unavailableText"`
}

func init() {
	c := &LicenseController{BaseController: controller.GetBaseController()}
	c.RegisterAdminApi("POST", "/license/page", c.Page)
	c.RegisterAdminApi("POST", "/license/abandon", c.Abandon)
	c.RegisterAdminApi("POST", "/license/export", c.Export)
	c.RegisterAdminApi("POST", "/license/import", c.Import)
}

func (c *LicenseController) Page(ctx *gin.Context) {
	provider := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider)
	view := officialLicenseSnapshotView{}
	if provider == nil {
		view.UnavailableText = "商业授权服务未初始化"
	} else if snapshot, err := provider.Current(ctx); err != nil {
		if errors.Is(err, official_license.ErrUnauthorized) {
			view.UnavailableText = "当前节点未安装有效插件商业授权"
		} else {
			view.UnavailableText = err.Error()
		}
	} else if snapshot != nil {
		view = buildOfficialLicenseSnapshotView(snapshot)
	}
	c.OkData(ctx, map[string]any{"license": view})
}

func (c *LicenseController) Abandon(ctx *gin.Context) {
	abandoner, ok := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider).(interface {
		AbandonLicense(ctx context.Context) error
	})
	if !ok {
		c.FailMsg(ctx, "当前商业授权 Provider 不支持放弃授权")
		return
	}
	if err := abandoner.AbandonLicense(ctx); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, true)
}

func (c *LicenseController) Export(ctx *gin.Context) {
	exporter, ok := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider).(interface {
		ExportLicense(ctx context.Context) (any, error)
	})
	if !ok {
		c.FailMsg(ctx, "当前商业授权 Provider 不支持导出")
		return
	}
	license, err := exporter.ExportLicense(ctx)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	data, err := json.MarshalIndent(license, "", "  ")
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	fileName := "插件授权信息.v9os"
	encodedFileName := url.QueryEscape(fileName)
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.Header("Content-Disposition", `attachment; filename="`+fileName+`"; filename*=UTF-8''`+encodedFileName)
	ctx.Header("File-Name", encodedFileName)
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Data(200, "application/json; charset=utf-8", data)
}

func (c *LicenseController) Import(ctx *gin.Context) {
	importer, ok := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider).(interface {
		ImportLicense(ctx context.Context, license any) error
	})
	if !ok {
		c.FailMsg(ctx, "当前商业授权 Provider 不支持导入")
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	src, err := file.Open()
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var license any
	if err := json.Unmarshal(data, &license); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	if err := importer.ImportLicense(ctx, license); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	provider := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider)
	snapshot, _ := provider.Current(ctx)
	c.OkData(ctx, snapshot)
}

func buildOfficialLicenseSnapshotView(snapshot *official_license.LicenseSnapshot) officialLicenseSnapshotView {
	features := make([]officialLicenseFeatureView, 0, len(snapshot.Features))
	codes := make([]string, 0, len(snapshot.Features))
	for code := range snapshot.Features {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	now := time.Now()
	for _, code := range codes {
		auth := snapshot.Features[code]
		item := officialLicenseFeatureView{
			Code:      code,
			Name:      auth.ProductName,
			AuthType:  auth.AuthType,
			Status:    "valid",
			HasCipher: auth.AuthCipher != "",
		}
		switch auth.AuthType {
		case "expired":
			if auth.Expired == nil {
				item.Status = "disabled"
				break
			}
			item.ExpiredAt = auth.Expired.Unix()
			if now.After(*auth.Expired) {
				item.Status = "expired"
			}
		case "times":
			item.Quantity = auth.Times
		case "has":
			item.Quantity = 1
		}
		if !auth.Has {
			item.Status = "disabled"
		}
		features = append(features, item)
	}
	return officialLicenseSnapshotView{
		Authorized:   true,
		AuthID:       snapshot.AuthID,
		StartAt:      snapshot.StartAt,
		EndAt:        snapshot.EndAt,
		FeatureCount: len(features),
		Features:     features,
	}
}
