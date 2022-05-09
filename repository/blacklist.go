package repository

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
)

// Blacklist ...
type Blacklist interface {
	GetBlacklistCustomers(ctx context.Context, phone string) ([]model.BlacklistCustomer, error)
}

type blacklistRepo struct {
}
