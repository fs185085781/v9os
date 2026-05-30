package official_license

import (
	"context"
	"time"
)

// FeatureAuth 是运行时使用的单个功能授权结果。
// 持久化 license 中的时间是秒级时间戳，注入内存后转换成 time.Time 方便判断。
type FeatureAuth struct {
	Has         bool       `json:"has"`
	AuthType    string     `json:"authType,omitempty"`
	ProductName string     `json:"productName,omitempty"`
	Expired     *time.Time `json:"expired,omitempty"`
	Times       int64      `json:"times,omitempty"`
	AuthCipher  string     `json:"authCipher,omitempty"`
}

// LicenseSnapshot 是本机 data.official_license 解密和校验通过后的授权快照。
// 其中 Features 已经是可直接用于 HasAuth 系列函数判断的内存 map。
type LicenseSnapshot struct {
	AuthID           string                 `json:"authId"`
	LastBatchID      string                 `json:"lastBatchId"`
	StartAt          int64                  `json:"startAt"`
	EndAt            int64                  `json:"endAt"`
	OfflineMonths    int64                  `json:"offlineMonths"`
	Features         map[string]FeatureAuth `json:"features"`
	AuthKey          string                 `json:"authKey"`
	LicensePublicKey string                 `json:"licensePublicKey"`
	PublicKeySign    string                 `json:"publicKeySign"`
	LastSyncAt       int64                  `json:"lastSyncAt"`
	NextSyncAt       int64                  `json:"nextSyncAt"`
	SyncError        string                 `json:"syncError"`
}

type LicenseProduct struct {
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AuthType    string              `json:"authType"`
	TargetType  string              `json:"targetType"`
	PublicKey   string              `json:"publicKey"`
	PriceGroups []LicensePriceGroup `json:"priceGroups"`
}

type LicensePriceGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Unit      string `json:"unit"`
	Quantity  int64  `json:"quantity"`
	PriceCent int64  `json:"priceCent"`
	Currency  string `json:"currency"`
}

type LicenseUpdateRequest struct {
	AuthID     string           `json:"authId"`
	OldLicense *LicenseSnapshot `json:"oldLicense"`
}

type LicenseUpdateResult struct {
	License any    `json:"license"`
	Note    string `json:"note"`
}

type LicensePurchaseRequest struct {
	AuthID        string                `json:"authId"`
	OldLicense    *LicenseSnapshot      `json:"oldLicense"`
	Items         []LicensePurchaseItem `json:"items"`
	OfflineMonths int64                 `json:"offlineMonths"`
	CouponCode    string                `json:"couponCode"`
	ReturnURL     string                `json:"returnUrl"`
	Force         bool                  `json:"force"`
}

type LicensePurchaseItem struct {
	ProductCode string `json:"productCode"`
	PriceID     string `json:"priceId"`
	Quantity    int64  `json:"quantity"`
}

type LicensePurchaseResult struct {
	PayID     string                `json:"payId"`
	PayURL    string                `json:"payUrl"`
	TotalCent int64                 `json:"totalCent"`
	Currency  string                `json:"currency"`
	Items     []LicensePurchaseLine `json:"items"`
	Note      string                `json:"note"`
}

type LicensePurchaseLine struct {
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
	PriceID     string `json:"priceId"`
	PriceName   string `json:"priceName"`
	Quantity    int64  `json:"quantity"`
	Unit        string `json:"unit"`
	UnitCent    int64  `json:"unitCent"`
	AmountCent  int64  `json:"amountCent"`
}

type LicensePaymentStatus struct {
	PayID   string `json:"payId"`
	Status  string `json:"status"`
	Note    string `json:"note"`
	License any    `json:"license"`
}

type LicenseAuthCipher struct {
	AuthID     string `json:"authId"`
	EndAt      int64  `json:"endAt"`
	AuthType   string `json:"authType,omitempty"`
	Value      string `json:"value,omitempty"`
	AuthCipher string `json:"authCipher"`
}

// Provider 是商业授权在内核中的唯一抽象。
// 插件侧 main/share 只会通过 pluginExp 桥接到这些检查函数。
type Provider interface {
	HasAuth(ctx context.Context, code string) error
	HasAuthByTimes(ctx context.Context, code string) (int64, error)
	AuthCipher(ctx context.Context, code string) (*LicenseAuthCipher, error)
	AllCommercial(ctx context.Context) bool
	Current(ctx context.Context) (*LicenseSnapshot, error)
	Sync(ctx context.Context) error
}
