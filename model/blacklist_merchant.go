package model

import (
	"database/sql"
	"time"
)

// BlacklistMerchant ...
type BlacklistMerchant struct {
	Hash         uint32                  `db:"hash"`
	MerchantCode string                  `db:"merchant_code"`
	Status       BlacklistMerchantStatus `db:"status"`
	StartTime    sql.NullTime            `db:"start_time"`
	EndTime      sql.NullTime            `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// NullBlacklistMerchant ...
type NullBlacklistMerchant struct {
	Valid    bool
	Merchant BlacklistMerchant
}

// BlacklistMerchantStatus ...
type BlacklistMerchantStatus int

const (
	// BlacklistMerchantStatusActive ...
	BlacklistMerchantStatusActive BlacklistMerchantStatus = 1

	// BlacklistMerchantStatusInactive ...
	BlacklistMerchantStatusInactive BlacklistMerchantStatus = 2
)
