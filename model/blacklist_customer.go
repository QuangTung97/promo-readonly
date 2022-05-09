package model

import (
	"database/sql"
	"time"
)

// BlacklistCustomer ...
type BlacklistCustomer struct {
	Hash      uint32                  `db:"hash"`
	Phone     string                  `db:"phone"`
	Status    BlacklistCustomerStatus `db:"status"`
	StartTime sql.NullTime            `db:"start_time"`
	EndTime   sql.NullTime            `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// NullBlacklistCustomer ...
type NullBlacklistCustomer struct {
	Valid    bool
	Customer BlacklistCustomer
}

// BlacklistCustomerStatus ...
type BlacklistCustomerStatus int

const (
	// BlacklistCustomerStatusActive ...
	BlacklistCustomerStatusActive BlacklistCustomerStatus = 1

	// BlacklistCustomerStatusInactive ...
	BlacklistCustomerStatusInactive BlacklistCustomerStatus = 2
)
