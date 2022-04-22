package model

import (
	"database/sql"
	"time"
)

// BlacklistMerchant ...
type BlacklistMerchant struct {
	Status       BlacklistMerchantStatus `db:"status"`
	Hash         uint32                  `db:"hash"`
	MerchantCode string                  `db:"merchant_code"`
	StartTime    sql.NullTime            `db:"start_time"`
	EndTime      sql.NullTime            `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// BlacklistMerchantStatus ...
type BlacklistMerchantStatus int

const (
	// BlacklistMerchantStatusActive ...
	BlacklistMerchantStatusActive BlacklistMerchantStatus = 1

	// BlacklistMerchantStatusInactive ...
	BlacklistMerchantStatusInactive BlacklistMerchantStatus = 2
)
