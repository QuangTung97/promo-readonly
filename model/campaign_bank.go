package model

import (
	"database/sql"
	"time"
)

// CampaignBank ...
type CampaignBank struct {
	CampaignID int64  `db:"campaign_id"`
	Hash       uint32 `db:"hash"`
	BankCode   string `db:"bank_code"`

	Status    CampaignBankStatus `db:"status"`
	StartTime sql.NullTime       `db:"start_time"`
	EndTime   sql.NullTime       `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CampaignBankStatus ...
type CampaignBankStatus int

const (
	//CampaignBankStatusActive ...
	CampaignBankStatusActive CampaignBankStatus = 1

	// CampaignBankStatusInactive ...
	CampaignBankStatusInactive CampaignBankStatus = 2
)
