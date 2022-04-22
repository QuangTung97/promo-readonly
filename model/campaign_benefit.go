package model

import (
	"github.com/shopspring/decimal"
	"time"
)

// CampaignBenefit ...
type CampaignBenefit struct {
	ID         int64     `db:"id"`
	CampaignID int64     `db:"campaign_id"`
	StartTime  time.Time `db:"start_time"`
	EndTime    time.Time `db:"end_time"`

	TxnMinAmount      decimal.Decimal `db:"txn_min_amount"`
	DiscountPercent   decimal.Decimal `db:"discount_percent"`
	MaxDiscountAmount decimal.Decimal `db:"max_discount_amount"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
