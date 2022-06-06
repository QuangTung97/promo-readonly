package readonly

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/util"
	"github.com/QuangTung97/promo-readonly/repository"
	"math/bits"
)

// IRepository ...
type IRepository interface {
	GetBlacklistCustomer(ctx context.Context, phone string) func() (model.NullBlacklistCustomer, error)
	GetBlacklistMerchant(ctx context.Context, merchantCode string) func() (model.NullBlacklistMerchant, error)
	// GetCampaigns(ctx context.Context, voucherCode string) func() ([]model.Campaign, error)
}

type repositoryImpl struct {
	sess                  dhash.Session
	blacklistCustomerHash dhash.Hash
	blacklistMerchantHash dhash.Hash
}

var _ IRepository = &repositoryImpl{}

// NewRepository ...
func NewRepository(sess dhash.Session, blacklistRepo repository.Blacklist) IRepository {
	return &repositoryImpl{
		sess: sess,

		blacklistCustomerHash: sess.NewHash("bl:cst", newBlacklistCustomerHashDB(blacklistRepo)),
		blacklistMerchantHash: sess.NewHash("bl:mc", newBlacklistMerchantHashDB(blacklistRepo)),
	}
}

func log2Int(n int64) uint64 {
	if n == 0 {
		return 0
	}
	return 64 - uint64(bits.LeadingZeros64(uint64(n-1)))
}

// GetBlacklistCustomer ...
func (r *repositoryImpl) GetBlacklistCustomer(
	ctx context.Context, phone string,
) func() (model.NullBlacklistCustomer, error) {
	hashValue := util.HashFunc(phone)
	fn := r.blacklistCustomerHash.SelectEntries(ctx, hashValue)
	return func() (model.NullBlacklistCustomer, error) {
		entries, err := fn()
		if err != nil {
			return model.NullBlacklistCustomer{}, err
		}
		for _, entry := range entries {
			if entry.Hash != hashValue {
				continue
			}

			customer, err := unmarshalBlacklistCustomer(entry.Data)
			if err != nil {
				return model.NullBlacklistCustomer{}, err
			}
			if customer.Phone != phone {
				continue
			}
			return model.NullBlacklistCustomer{
				Valid:    true,
				Customer: customer,
			}, nil
		}
		return model.NullBlacklistCustomer{}, nil
	}
}

// GetBlacklistMerchant ...
func (r *repositoryImpl) GetBlacklistMerchant(
	ctx context.Context, merchantCode string,
) func() (model.NullBlacklistMerchant, error) {
	hashValue := util.HashFunc(merchantCode)
	fn := r.blacklistMerchantHash.SelectEntries(ctx, hashValue)
	return func() (model.NullBlacklistMerchant, error) {
		entries, err := fn()
		if err != nil {
			return model.NullBlacklistMerchant{}, err
		}
		for _, entry := range entries {
			if entry.Hash != hashValue {
				continue
			}

			merchant, err := unmarshalBlacklistMerchant(entry.Data)
			if err != nil {
				return model.NullBlacklistMerchant{}, err
			}
			if merchant.MerchantCode != merchantCode {
				continue
			}
			return model.NullBlacklistMerchant{
				Valid:    true,
				Merchant: merchant,
			}, nil
		}
		return model.NullBlacklistMerchant{}, nil
	}
}
