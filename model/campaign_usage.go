package model

import (
	"github.com/shopspring/decimal"
	"time"
)

// CampaignUsage ...
type CampaignUsage struct {
	CampaignID   int64           `db:"campaign_id"`
	BudgetUsed   decimal.Decimal `db:"budget_used"`
	CampaignUsed int64           `db:"campaign_used"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
