package model

// Campaign ...
type Campaign struct {
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Status      CampaignStatus `db:"status"`
	VoucherHash uint32         `db:"voucher_hash"`
	VoucherCode string         `db:"voucher_code"`
}

// CampaignStatus ...
type CampaignStatus int

const (
	// CampaignStatusActive ...
	CampaignStatusActive CampaignStatus = 1

	// CampaignStatusInactive ...
	CampaignStatusInactive CampaignStatus = 2
)
