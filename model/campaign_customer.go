package model

import (
	"database/sql"
	"time"
)

// CampaignCustomer ...
type CampaignCustomer struct {
	CampaignID int64 `db:"campaign_id"`

	Hash      uint32                 `db:"hash"`
	Phone     string                 `db:"phone"`
	Status    CampaignCustomerStatus `db:"status"`
	StartTime sql.NullTime           `db:"start_time"`
	EndTime   sql.NullTime           `db:"end_time"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CampaignCustomerStatus ...
type CampaignCustomerStatus int

const (
	//CampaignCustomerStatusActive ...
	CampaignCustomerStatusActive CampaignCustomerStatus = 1

	// CampaignCustomerStatusInactive ...
	CampaignCustomerStatusInactive CampaignCustomerStatus = 2
)
