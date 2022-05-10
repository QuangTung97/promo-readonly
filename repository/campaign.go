package repository

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
	"time"
)

// Campaign ...
type Campaign interface {
	FindCampaignsByVoucher(
		ctx context.Context, voucherHash uint32, voucherCode string, now time.Time,
	) ([]model.Campaign, error)
	LockCampaign(ctx context.Context, campaignID int64) error
	UpsertCampaign(ctx context.Context, campaign model.Campaign) error
}

type campaignImpl struct {
}

// NewCampaign ...
func NewCampaign() Campaign {
	return &campaignImpl{}
}

// FindCampaignsByVoucher ...
func (c *campaignImpl) FindCampaignsByVoucher(
	ctx context.Context, voucherHash uint32, voucherCode string, now time.Time,
) ([]model.Campaign, error) {
	query := `
SELECT id, name, status, type, voucher_hash, voucher_code, start_time, end_time,
	budget_max, campaign_usage_max, customer_usage_max,
	period_usage_type, period_customer_usage_max, period_term_type,
	all_merchants
FROM campaign
WHERE voucher_hash = ? AND voucher_code = ? AND ? < end_time
`
	var result []model.Campaign
	err := GetReadonly(ctx).SelectContext(ctx, &result, query, voucherHash, voucherCode, now)
	return result, err
}

// LockCampaign ...
func (c *campaignImpl) LockCampaign(ctx context.Context, campaignID int64) error {
	query := `SELECT id FROM campaign WHERE id = ? FOR UPDATE`
	var id int64
	return GetTx(ctx).GetContext(ctx, &id, query, campaignID)
}

// UpsertCampaign ...
func (c *campaignImpl) UpsertCampaign(ctx context.Context, campaign model.Campaign) error {
	query := `
INSERT INTO campaign (
	id, name, status, type,
	voucher_hash, voucher_code, start_time, end_time,
	budget_max, campaign_usage_max, customer_usage_max,
	period_usage_type, period_customer_usage_max, period_term_type,
	all_merchants
) VALUES (
	:id, :name, :status, :type,
	:voucher_hash, :voucher_code, :start_time, :end_time,
	:budget_max, :campaign_usage_max, :customer_usage_max,
	:period_usage_type, :period_customer_usage_max, :period_term_type,
	:all_merchants
) AS NEW
ON DUPLICATE KEY UPDATE
	name = NEW.name,
	status = NEW.status,
	voucher_hash = NEW.voucher_hash,
	voucher_code = NEW.voucher_code,
	start_time = NEW.start_time,
	end_time = NEW.end_time,

	budget_max = NEW.budget_max,
	campaign_usage_max = NEW.campaign_usage_max,
	customer_usage_max = NEW.customer_usage_max,

	period_usage_type = NEW.period_usage_type,
	period_customer_usage_max = NEW.period_customer_usage_max,
	period_term_type = NEW.period_term_type,

	all_merchants = NEW.all_merchants
`
	_, err := GetTx(ctx).NamedExecContext(ctx, query, campaign)
	return err
}
