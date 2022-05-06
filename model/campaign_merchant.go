package model

import (
	"database/sql"
	"time"
)

// CampaignMerchant ...
type CampaignMerchant struct {
	CampaignID   int64  `db:"campaign_id"`
	Hash         uint32 `db:"hash"`
	MerchantCode string `db:"merchant_code"`

	Status       CampaignMerchantStatus `db:"status"`
	StartTime    sql.NullTime           `db:"start_time"`
	EndTime      sql.NullTime           `db:"end_time"`
	AllTerminals bool                   `db:"all_terminals"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CampaignMerchantStatus ...
type CampaignMerchantStatus int

const (
	// CampaignMerchantStatusActive ...
	CampaignMerchantStatusActive CampaignMerchantStatus = 1

	// CampaignMerchantStatusInactive ...
	CampaignMerchantStatusInactive CampaignMerchantStatus = 2
)
