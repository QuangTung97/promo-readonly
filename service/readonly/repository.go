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
	// GetCampaigns(ctx context.Context, voucherCode string) func() ([]model.Campaign, error)
}

type repositoryImpl struct {
	sess          dhash.Session
	blacklistRepo repository.Blacklist
}

var _ IRepository = &repositoryImpl{}

// NewRepository ...
func NewRepository(sess dhash.Session, blacklistRepo repository.Blacklist) IRepository {
	return &repositoryImpl{
		sess:          sess,
		blacklistRepo: blacklistRepo,
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
	hash := r.sess.NewHash("bl:cst", newBlacklistCustomerHashDB(r.blacklistRepo))
	fn := hash.SelectEntries(ctx, hashValue)
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
