package repository

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/model"
	"strings"
)

// BlacklistCustomerKey ...
type BlacklistCustomerKey struct {
	Hash  uint32
	Phone string
}

// Blacklist ...
type Blacklist interface {
	GetBlacklistCustomers(ctx context.Context, keys []BlacklistCustomerKey) ([]model.BlacklistCustomer, error)
	UpsertBlacklistCustomers(ctx context.Context, customers []model.BlacklistCustomer) error
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
