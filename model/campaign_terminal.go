package model

import (
	"database/sql"
	"time"
)

// CampaignTerminal ...
type CampaignTerminal struct {
	CampaignID   int64  `db:"campaign_id"`
	Hash         uint32 `db:"hash"`
	MerchantCode string `db:"merchant_code"`
	TerminalCode string `db:"terminal_code"`

	Status    CampaignTerminalStatus `db:"status"`
	StartTime sql.NullTime           `db:"start_time"`
	EndTime   sql.NullTime           `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CampaignTerminalStatus ...
type CampaignTerminalStatus int

const (
	// CampaignTerminalStatusActive ...
	CampaignTerminalStatusActive CampaignTerminalStatus = 1

	// CampaignTerminalStatusInactive ...
	CampaignTerminalStatusInactive CampaignTerminalStatus = 2
)
