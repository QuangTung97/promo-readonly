package model

import "time"

// BlacklistConfig ...
type BlacklistConfig struct {
	ID            int64 `db:"id"`
	CustomerCount int64 `db:"customer_count"`
	MerchantCount int64 `db:"merchant_count"`
	TerminalCount int64 `db:"terminal_count"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
