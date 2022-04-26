package model

import "time"

// CampaignPeriodUsage ...
type CampaignPeriodUsage struct {
	CampaignID int64  `db:"campaign_id"`
	Hash       uint32 `db:"hash"`
	Phone      string `db:"phone"`
	TermCode   string `db:"term_code"`

	UsageNum  int64     `db:"usage_num"`
	ExpiredOn time.Time `db:"expired_on"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
