package model

import "time"

// BlacklistCustomer ...
type BlacklistCustomer struct {
	Hash      uint32                  `db:"hash"`
	Phone     string                  `db:"phone"`
	Status    BlacklistCustomerStatus `db:"status"`
	StartTime time.Time               `db:"start_time"`
	EndTime   time.Time               `db:"end_time"`

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
