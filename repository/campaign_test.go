package repository

import (
	"context"
	"database/sql"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/integration"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

type campaignTest struct {
	tc       *integration.TestCase
	provider Provider
}

func newCampaignTest() *campaignTest {
	tc := integration.NewTestCase()
	return &campaignTest{
		tc:       tc,
		provider: NewProvider(tc.DB),
	}
}

func newDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func newNullDecimal(s string) decimal.NullDecimal {
	return decimal.NewNullDecimal(newDecimal(s))
}

func newNullInt64(n int64) sql.NullInt64 {
	return sql.NullInt64{Valid: true, Int64: n}
}

func TestCampaign(t *testing.T) {
	tc := newCampaignTest()
	tc.tc.Truncate("campaign")

	repo := NewCampaign()

	ctx := tc.provider.Readonly(newContext())

	const hash01 = 3300
	const voucherCode01 = "VOUCHER01"

	// Get 1
	campaigns, err := repo.FindCampaignsByVoucher(ctx,
		hash01, voucherCode01, newTime("2022-05-13T10:00:00+07:00"))
	assert.Equal(t, nil, err)
	assert.Nil(t, campaigns)

	campaign01 := model.Campaign{
		Name:   "name 01",
		Status: model.CampaignStatusActive,
		Type:   model.CampaignTypeMerchant,

		VoucherHash: hash01,
		VoucherCode: voucherCode01,
		StartTime:   newTime("2022-05-07T10:00:00+07:00"),
		EndTime:     newTime("2022-05-14T10:00:00+07:00"),

		BudgetMax:        newNullDecimal("120000.00"),
		CampaignUsageMax: newNullInt64(200),
		CustomerUsageMax: 5,

		PeriodUsageType:        model.PeriodUsageTypeDaily,
		PeriodCustomerUsageMax: newNullInt64(2),
		PeriodTermType:         model.PeriodTermTypeCampaign,

		AllMerchants: true,
	}

	// Insert
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertCampaign(ctx, campaign01)
	})
	assert.Equal(t, nil, err)

	// Get 2
	campaigns, err = repo.FindCampaignsByVoucher(ctx,
		hash01, voucherCode01, newTime("2022-05-13T10:00:00+07:00"))
	assert.Equal(t, nil, err)

	campaign01.ID = 1
	assert.Equal(t, []model.Campaign{campaign01}, campaigns)

	// Lock Campaign
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.LockCampaign(ctx, 1)
	})
	assert.Equal(t, nil, err)

	// Upsert
	campaign01.Name = "name 02"
	campaign01.Status = model.CampaignStatusInactive
	campaign01.VoucherHash = 4400
	campaign01.VoucherCode = "VOUCHER02"
	campaign01.StartTime = newTime("2022-07-07T10:00:00+07:00")
	campaign01.EndTime = newTime("2022-07-14T10:00:00+07:00")

	campaign01.BudgetMax = newNullDecimal("88000.00")
	campaign01.CampaignUsageMax = newNullInt64(90)
	campaign01.CustomerUsageMax = 12

	campaign01.PeriodUsageType = model.PeriodUsageTypeMonthly
	campaign01.PeriodCustomerUsageMax = newNullInt64(4)
	campaign01.PeriodTermType = model.PeriodTermTypeTerminal

	campaign01.AllMerchants = false

	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertCampaign(ctx, campaign01)
	})
	assert.Equal(t, nil, err)

	// Get 3
	campaigns, err = repo.FindCampaignsByVoucher(ctx,
		4400, "VOUCHER02", newTime("2022-07-13T10:00:00+07:00"))
	assert.Equal(t, nil, err)
	assert.Equal(t, []model.Campaign{campaign01}, campaigns)
}
