package readonly

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
)

// IRepository ...
type IRepository interface {
	GetBlacklistCustomer(ctx context.Context, phone string) func(customer model.BlacklistCustomer)
	GetCampaigns(ctx context.Context, voucherCode string) func() ([]model.Campaign, error)
}
