package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"strings"
)

// HashRange ...
type HashRange struct {
	Begin uint32
	End   dhash.NullUint32
}

//go:generate moq -rm -out blacklist_mocks.go . Blacklist

// Blacklist ...
type Blacklist interface {
	GetConfig(ctx context.Context) (model.BlacklistConfig, error)
	UpsertConfig(ctx context.Context, config model.BlacklistConfig) error

	GetBlacklistCustomers(ctx context.Context, keys []BlacklistCustomerKey) ([]model.BlacklistCustomer, error)
	SelectBlacklistCustomers(ctx context.Context, ranges []HashRange) ([]model.BlacklistCustomer, error)
	UpsertBlacklistCustomers(ctx context.Context, customers []model.BlacklistCustomer) error

	GetBlacklistMerchants(ctx context.Context, keys []BlacklistMerchantKey) ([]model.BlacklistMerchant, error)
	UpsertBlacklistMerchants(ctx context.Context, merchants []model.BlacklistMerchant) error

	GetBlacklistTerminals(ctx context.Context, keys []BlacklistTerminalKey) ([]model.BlacklistTerminal, error)
	UpsertBlacklistTerminals(ctx context.Context, terminals []model.BlacklistTerminal) error
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

// BlacklistTerminalKey ...
type BlacklistTerminalKey struct {
	Hash         uint32
	MerchantCode string
	TerminalCode string
}

type blacklistRepo struct {
}

var _ Blacklist = &blacklistRepo{}

// NewBlacklist ...
func NewBlacklist() Blacklist {
	return &blacklistRepo{}
}

// GetConfig ...
func (b *blacklistRepo) GetConfig(ctx context.Context) (model.BlacklistConfig, error) {
	query := `
SELECT id, customer_count, merchant_count, terminal_count
FROM blacklist_config WHERE id = 1
`
	var result model.BlacklistConfig
	err := GetReadonly(ctx).GetContext(ctx, &result, query)
	if err == sql.ErrNoRows {
		return model.BlacklistConfig{}, nil
	}
	return result, err
}

// UpsertConfig ...
func (b *blacklistRepo) UpsertConfig(ctx context.Context, config model.BlacklistConfig) error {
	config.ID = 1
	query := `
INSERT INTO blacklist_config (id, customer_count, merchant_count, terminal_count)
VALUES (:id, :customer_count, :merchant_count, :terminal_count) AS NEW
ON DUPLICATE KEY UPDATE
	customer_count = NEW.customer_count,
	merchant_count = NEW.merchant_count,
	terminal_count = NEW.terminal_count
`
	_, err := GetTx(ctx).NamedExecContext(ctx, query, config)
	return err
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

func (b *blacklistRepo) SelectBlacklistCustomers(
	ctx context.Context, ranges []HashRange,
) ([]model.BlacklistCustomer, error) {
	if len(ranges) == 0 {
		return nil, nil
	}

	var buf strings.Builder
	query := `
SELECT hash, phone, status, start_time, end_time
FROM blacklist_customer WHERE hash >= ?%s
`

	withEndQuery := fmt.Sprintf(query, " AND hash < ?")
	noEndQuery := fmt.Sprintf(query, "")

	args := make([]interface{}, 0, 2*len(ranges))

	for i, r := range ranges {
		if i > 0 {
			buf.WriteString("UNION ALL")
		}
		args = append(args, r.Begin)

		if r.End.Valid {
			buf.WriteString(withEndQuery)
			args = append(args, r.End.Num)
		} else {
			buf.WriteString(noEndQuery)
		}
	}

	var result []model.BlacklistCustomer
	err := GetReadonly(ctx).SelectContext(ctx, &result, buf.String(), args...)
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

// GetBlacklistTerminals ...
func (b *blacklistRepo) GetBlacklistTerminals(
	ctx context.Context, keys []BlacklistTerminalKey,
) ([]model.BlacklistTerminal, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var buf strings.Builder
	const placeholder = "(?, ?, ?)"
	buf.WriteString(placeholder)
	for range keys[1:] {
		buf.WriteString("," + placeholder)
	}

	query := fmt.Sprintf(`
SELECT hash, merchant_code, terminal_code, status, start_time, end_time
FROM blacklist_terminal WHERE (hash, merchant_code, terminal_code) IN (%s)
`, buf.String())

	args := make([]interface{}, 0, 3*len(keys))
	for _, key := range keys {
		args = append(args, key.Hash, key.MerchantCode, key.TerminalCode)
	}

	var result []model.BlacklistTerminal
	err := GetReadonly(ctx).SelectContext(ctx, &result, query, args...)
	return result, err
}

// UpsertBlacklistTerminals ...
func (b *blacklistRepo) UpsertBlacklistTerminals(ctx context.Context, terminals []model.BlacklistTerminal) error {
	if len(terminals) == 0 {
		return nil
	}

	query := `
INSERT INTO blacklist_terminal (hash, merchant_code, terminal_code, status, start_time, end_time)
VALUES (:hash, :merchant_code, :terminal_code, :status, :start_time, :end_time) AS NEW
ON DUPLICATE KEY UPDATE
	status = NEW.status,
	start_time = NEW.start_time,
	end_time = NEW.end_time
`
	_, err := GetTx(ctx).NamedExecContext(ctx, query, terminals)
	return err
}
