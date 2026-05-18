package official_license

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
)

var ErrUnauthorized = errors.New("official license unauthorized")

const defaultDataKey = "official_license"

type DefaultProvider struct {
	mu       sync.RWMutex
	loaded   bool
	snapshot *LicenseSnapshot
}

type defaultLicenseJSON struct {
	AuthID           string                        `json:"authId"`
	LastBatchID      string                        `json:"lastBatchId"`
	StartAt          int64                         `json:"startAt"`
	EndAt            int64                         `json:"endAt"`
	OfflineMonths    int64                         `json:"offlineMonths"`
	Features         map[string]defaultFeatureJSON `json:"features"`
	AuthKey          string                        `json:"authKey"`
	LicensePublicKey string                        `json:"licensePublicKey"`
	PublicKeySign    string                        `json:"publicKeySign"`
	LastSyncAt       int64                         `json:"lastSyncAt"`
	NextSyncAt       int64                         `json:"nextSyncAt"`
	SyncError        string                        `json:"syncError"`
}

type defaultFeatureJSON struct {
	Has         bool   `json:"has"`
	AuthType    string `json:"authType"`
	ProductName string `json:"productName,omitempty"`
	Expired     int64  `json:"expired"`
	Times       int64  `json:"times"`
	PublicKey   string `json:"publicKey"`
	AuthCipher  string `json:"authCipher"`
}

func (d *DefaultProvider) HasAuth(ctx context.Context, code string) error {
	return nil
}

func (d *DefaultProvider) HasAuthByTimes(ctx context.Context, code string) (int64, error) {
	return math.MaxInt64, nil
}

func (d *DefaultProvider) AuthCipher(ctx context.Context, code string) (*LicenseAuthCipher, error) {
	snapshot, err := d.Current(ctx)
	if err != nil {
		return nil, err
	}
	auth, ok := snapshot.Features[strings.TrimSpace(code)]
	if !ok {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(auth.PublicKey) == "" || strings.TrimSpace(auth.AuthCipher) == "" {
		return nil, ErrUnauthorized
	}
	if auth.AuthType != "has" && auth.AuthType != "times" && auth.AuthType != "expired" {
		return nil, ErrUnauthorized
	}
	return &LicenseAuthCipher{
		AuthID:     snapshot.AuthID,
		AuthCipher: auth.AuthCipher,
	}, nil
}

func (d *DefaultProvider) AllCommercial(ctx context.Context) bool {
	return false
}

func (d *DefaultProvider) Current(ctx context.Context) (*LicenseSnapshot, error) {
	if err := d.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.snapshot == nil {
		return nil, ErrUnauthorized
	}
	return cloneDefaultSnapshot(d.snapshot), nil
}

func (d *DefaultProvider) Sync(ctx context.Context) error {
	return nil
}

func (d *DefaultProvider) ImportLicense(ctx context.Context, license any) error {
	raw, err := defaultLicenseFromAny(license)
	if err != nil {
		return err
	}
	snapshot := raw.snapshot()
	if snapshot.AuthID == "" || len(snapshot.Features) == 0 {
		return ErrUnauthorized
	}
	if err := uioc.Database().SetObject(defaultDataKey, raw); err != nil {
		return err
	}
	d.mu.Lock()
	d.loaded = true
	d.snapshot = snapshot
	d.mu.Unlock()
	return nil
}

func (d *DefaultProvider) ExportLicense(ctx context.Context) (any, error) {
	snapshot, err := d.Current(ctx)
	if err != nil {
		return nil, err
	}
	return defaultLicenseFromSnapshot(snapshot), nil
}

func (d *DefaultProvider) AbandonLicense(ctx context.Context) error {
	if err := uioc.Database().SetObject(defaultDataKey, map[string]any{}); err != nil {
		return err
	}
	d.mu.Lock()
	d.loaded = true
	d.snapshot = nil
	d.mu.Unlock()
	return nil
}

func (d *DefaultProvider) ensureLoaded(ctx context.Context) error {
	d.mu.RLock()
	if d.loaded {
		d.mu.RUnlock()
		return nil
	}
	d.mu.RUnlock()

	var raw defaultLicenseJSON
	has, err := uioc.Database().GetObject(defaultDataKey, &raw)
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.loaded = true
	if !has || strings.TrimSpace(raw.AuthID) == "" {
		d.snapshot = nil
		return ErrUnauthorized
	}
	d.snapshot = raw.snapshot()
	return nil
}

func (l defaultLicenseJSON) snapshot() *LicenseSnapshot {
	features := make(map[string]FeatureAuth, len(l.Features))
	for code, item := range l.Features {
		auth := FeatureAuth{
			Has:         item.Has,
			AuthType:    item.AuthType,
			ProductName: item.ProductName,
			Times:       item.Times,
			PublicKey:   item.PublicKey,
			AuthCipher:  item.AuthCipher,
		}
		if item.Expired > 0 {
			t := time.Unix(item.Expired, 0)
			auth.Expired = &t
		}
		switch item.AuthType {
		case "has":
			auth.Has = true
		case "times":
			auth.Has = item.Times > 0
		case "expired":
			auth.Has = item.Expired > 0
		}
		features[code] = auth
	}
	return &LicenseSnapshot{
		AuthID:        l.AuthID,
		LastBatchID:   l.LastBatchID,
		StartAt:       l.StartAt,
		EndAt:         l.EndAt,
		OfflineMonths: l.OfflineMonths,
		Features:      features,
		AuthKey:       l.AuthKey,
		LicensePublicKey: l.LicensePublicKey,
		PublicKeySign: l.PublicKeySign,
		LastSyncAt:    l.LastSyncAt,
		NextSyncAt:    l.NextSyncAt,
		SyncError:     l.SyncError,
	}
}

func defaultLicenseFromAny(value any) (*defaultLicenseJSON, error) {
	raw := &defaultLicenseJSON{}
	if value == nil {
		return nil, ErrUnauthorized
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func defaultLicenseFromSnapshot(snapshot *LicenseSnapshot) defaultLicenseJSON {
	features := make(map[string]defaultFeatureJSON, len(snapshot.Features))
	for code, item := range snapshot.Features {
		expired := int64(0)
		if item.Expired != nil {
			expired = item.Expired.Unix()
		}
		features[code] = defaultFeatureJSON{
			Has:         item.Has,
			AuthType:    item.AuthType,
			ProductName: item.ProductName,
			Expired:     expired,
			Times:       item.Times,
			PublicKey:   item.PublicKey,
			AuthCipher:  item.AuthCipher,
		}
	}
	return defaultLicenseJSON{
		AuthID:        snapshot.AuthID,
		LastBatchID:   snapshot.LastBatchID,
		StartAt:       snapshot.StartAt,
		EndAt:         snapshot.EndAt,
		OfflineMonths: snapshot.OfflineMonths,
		Features:      features,
		AuthKey:       snapshot.AuthKey,
		LicensePublicKey: snapshot.LicensePublicKey,
		PublicKeySign: snapshot.PublicKeySign,
		LastSyncAt:    snapshot.LastSyncAt,
		NextSyncAt:    snapshot.NextSyncAt,
		SyncError:     snapshot.SyncError,
	}
}

func cloneDefaultSnapshot(snapshot *LicenseSnapshot) *LicenseSnapshot {
	if snapshot == nil {
		return nil
	}
	next := *snapshot
	if snapshot.Features != nil {
		next.Features = make(map[string]FeatureAuth, len(snapshot.Features))
		for code, item := range snapshot.Features {
			copied := item
			if item.Expired != nil {
				t := *item.Expired
				copied.Expired = &t
			}
			next.Features[code] = copied
		}
	}
	return &next
}

func initProvider() {
	if ioc.Ioc().Get(ioc.KeyOfficialLicenseProvider) == nil {
		ioc.Ioc().Register(ioc.KeyOfficialLicenseProvider, &DefaultProvider{})
	}
}

func init() {
	initProvider()
}

var _ Provider = (*DefaultProvider)(nil)
