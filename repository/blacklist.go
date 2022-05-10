package repository

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/model"
	"strings"
)

// Blacklist ...
type Blacklist interface {
	GetBlacklistCustomers(ctx context.Context, keys []BlacklistCustomerKey) ([]model.BlacklistCustomer, error)
	UpsertBlacklistCustomers(ctx context.Context, customers []model.BlacklistCustomer) error

	GetBlacklistMerchants(ctx context.Context, keys []BlacklistMerchantKey) ([]model.BlacklistMerchant, error)
	UpsertBlacklistMerchants(ctx context.Context, merchants []model.BlacklistMerchant) error
}

// BlacklistCustomerKey ...
type BlacklistCustomerKey struct {
	Hash  uint32
	Phone string
}

// BlacklistMerchantKey ...
type BlacklistMerchantKey struct {
	Hash         uint32
	MerchantCode string
}

type blacklistRepo struct {
}

var _ Blacklist = &blacklistRepo{}

// NewBlacklist ...
func NewBlacklist() Blacklist {
	return &blacklistRepo{}
}

// GetBlacklistCustomers ...
func (b *blacklistRepo) GetBlacklistCustomers(
	ctx context.Context, keys []BlacklistCustomerKey,
) ([]model.BlacklistCustomer, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	const placeholder = "(?, ?)"
	var buf strings.Builder
	buf.WriteString(placeholder)
	for range keys[1:] {
		buf.WriteString("," + placeholder)
	}

	query := fmt.Sprintf(`
SELECT hash, phone, status, start_time, end_time
FROM blacklist_customer WHERE (hash, phone) IN (%s)
`, buf.String())

	args := make([]interface{}, 0, 2*len(keys))
	for _, key := range keys {
		args = append(args, key.Hash, key.Phone)
	}

	db := GetReadonly(ctx)
	var result []model.BlacklistCustomer
	err := db.SelectContext(ctx, &result, query, args...)
	return result, err
}

// UpsertBlacklistCustomers ...
func (b *blacklistRepo) UpsertBlacklistCustomers(ctx context.Context, customers []model.BlacklistCustomer) error {
	if len(customers) == 0 {
		return nil
	}

	query := `
INSERT INTO blacklist_customer (hash, phone, status, start_time, end_time)
VALUES (:hash, :phone, :status, :start_time, :end_time) AS NEW
ON DUPLICATE KEY UPDATE
	status = NEW.status,
	start_time = NEW.start_time,
	end_time = NEW.end_time
`
	tx := GetTx(ctx)
	_, err := tx.NamedExecContext(ctx, query, customers)
	return err
}

// GetBlacklistMerchants ...
func (b *blacklistRepo) GetBlacklistMerchants(
	ctx context.Context, keys []BlacklistMerchantKey,
) ([]model.BlacklistMerchant, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var buf strings.Builder
	const placeholder = "(?, ?)"
	buf.WriteString(placeholder)
	for range keys[1:] {
		buf.WriteString("," + placeholder)
	}

	query := fmt.Sprintf(`
SELECT hash, merchant_code, status, start_time, end_time
FROM blacklist_merchant WHERE (hash, merchant_code) IN (%s)
`, buf.String())

	args := make([]interface{}, 0, 2*len(keys))
	for _, key := range keys {
		args = append(args, key.Hash, key.MerchantCode)
	}

	var result []model.BlacklistMerchant
	err := GetReadonly(ctx).SelectContext(ctx, &result, query, args...)
	return result, err
}

// UpsertBlacklistMerchants ...
func (b *blacklistRepo) UpsertBlacklistMerchants(ctx context.Context, merchants []model.BlacklistMerchant) error {
	if len(merchants) == 0 {
		return nil
	}

	query := `
INSERT INTO blacklist_merchant (hash, merchant_code, status, start_time, end_time)
VALUES (:hash, :merchant_code, :status, :start_time, :end_time) AS NEW
ON DUPLICATE KEY UPDATE
	status = NEW.status,
	start_time = NEW.start_time,
	end_time = NEW.end_time
`
	_, err := GetTx(ctx).NamedExecContext(ctx, query, merchants)
	return err
}
