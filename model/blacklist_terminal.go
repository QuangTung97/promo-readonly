package model

import (
	"database/sql"
	"time"
)

// BlacklistTerminal ...
type BlacklistTerminal struct {
	Status       BlacklistTerminalStatus `db:"status"`
	Hash         uint32                  `db:"hash"`
	MerchantCode string                  `db:"merchant_code"`
	TerminalCode string                  `db:"terminal_code"`
	StartTime    sql.NullTime            `db:"start_time"`
	EndTime      sql.NullTime            `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// BlacklistTerminalStatus ...
type BlacklistTerminalStatus int

const (
	// BlacklistTerminalStatusActive ...
	BlacklistTerminalStatusActive BlacklistTerminalStatus = 1

	// BlacklistTerminalStatusInactive ...
	BlacklistTerminalStatusInactive BlacklistTerminalStatus = 2
)
