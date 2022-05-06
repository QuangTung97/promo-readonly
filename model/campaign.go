package model

import (
	"database/sql"
	"github.com/shopspring/decimal"
	"time"
)

// Campaign ...
type Campaign struct {
	ID     int64          `db:"id"`
	Name   string         `db:"name"`
	Status CampaignStatus `db:"status"`
	Type   CampaignType   `db:"type"`

	VoucherHash uint32    `db:"voucher_hash"`
	VoucherCode string    `db:"voucher_code"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`

	BudgetMax        decimal.NullDecimal `db:"budget_max"`
	CampaignUsageMax sql.NullInt64       `db:"campaign_usage_max"`
	CustomerUsageMax int64               `db:"customer_usage_max"`

	PeriodUsageType        PeriodUsageType `db:"period_usage_type"`
	PeriodCustomerUsageMax sql.NullInt64   `db:"period_customer_usage_max"`
	PeriodTermType         PeriodTermType  `db:"period_term_type"`

	AllMerchants bool `db:"all_merchants"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// CampaignStatus ...
type CampaignStatus int

const (
	// CampaignStatusActive ...
	CampaignStatusActive CampaignStatus = 1

	// CampaignStatusInactive ...
	CampaignStatusInactive CampaignStatus = 2
)

// CampaignType ...
type CampaignType int

const (
	// CampaignTypeMerchant ...
	CampaignTypeMerchant CampaignType = 1

	// CampaignTypeBank ...
	CampaignTypeBank CampaignType = 2

	// CampaignTypePrivate ...
	CampaignTypePrivate CampaignType = 3
)

// PeriodUsageType ...
type PeriodUsageType int

const (
	// PeriodUsageTypeUnspecified ...
	PeriodUsageTypeUnspecified PeriodUsageType = 0

	// PeriodUsageTypeDaily ...
	PeriodUsageTypeDaily PeriodUsageType = 1

	// PeriodUsageTypeWeekly ...
	PeriodUsageTypeWeekly PeriodUsageType = 2

	// PeriodUsageTypeMonthly ...
	PeriodUsageTypeMonthly PeriodUsageType = 3
)

// PeriodTermType ...
type PeriodTermType int

const (
	// PeriodTermTypeCampaign ...
	PeriodTermTypeCampaign PeriodTermType = 1

	// PeriodTermTypeMerchant ...
	PeriodTermTypeMerchant PeriodTermType = 2

	// PeriodTermTypeTerminal ...
	PeriodTermTypeTerminal PeriodTermType = 3
)
