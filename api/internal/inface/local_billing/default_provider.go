package local_billing

import (
	"context"

	"github.com/fs185085781/v9os/internal/ioc"
)

type DefaultProvider struct {
}

func (d *DefaultProvider) Status(ctx context.Context) Status {
	return Status{}
}

func initProvider() {
	if ioc.Ioc().Get(ioc.KeyLocalBillingProvider) != nil {
		return
	}
	ioc.Ioc().Register(ioc.KeyLocalBillingProvider, &DefaultProvider{})
}

func init() {
	initProvider()
}

var _ Provider = (*DefaultProvider)(nil)
