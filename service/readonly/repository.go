package readonly

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/util"
	"github.com/QuangTung97/promo-readonly/repository"
	"math/bits"
	"time"
)

// IRepositoryProvider ...
type IRepositoryProvider interface {
	NewRepo() IRepository
}

// IRepository ...
type IRepository interface {
	GetBlacklistCustomer(ctx context.Context, phone string) func() (model.NullBlacklistCustomer, error)
	GetBlacklistMerchant(ctx context.Context, merchantCode string) func() (model.NullBlacklistMerchant, error)
	// GetCampaigns(ctx context.Context, voucherCode string) func() ([]model.Campaign, error)

	Finish()
}

type repositoryProviderImpl struct {
	dhashProvider dhash.Provider
	blacklistRepo repository.Blacklist
}

var _ IRepositoryProvider = &repositoryProviderImpl{}

// NewRepositoryProvider ...
func NewRepositoryProvider(provider dhash.Provider, blacklistRepo repository.Blacklist) IRepositoryProvider {
	return &repositoryProviderImpl{
		dhashProvider: provider,
		blacklistRepo: blacklistRepo,
	}
}

// NewRepo ...
func (p *repositoryProviderImpl) NewRepo() IRepository {
	sess := p.dhashProvider.NewSession(dhash.WithWaitLeaseDurations([]time.Duration{
		4 * time.Millisecond,
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
	}))

	return newRepository(sess,
		sess.NewHash("bl:cst", newBlacklistCustomerHashDB(p.blacklistRepo)),
		sess.NewHash("bl:mc", newBlacklistMerchantHashDB(p.blacklistRepo)),
	)
}

func newRepository(
	sess dhash.Session, blacklistCustomerHash dhash.Hash, blacklistMerchantHash dhash.Hash,
) IRepository {
	return &repositoryImpl{
		sess: sess,

		blacklistCustomerHash: blacklistCustomerHash,
		blacklistMerchantHash: blacklistMerchantHash,
	}
}

type repositoryImpl struct {
	sess                  dhash.Session
	blacklistCustomerHash dhash.Hash
	blacklistMerchantHash dhash.Hash
}

var _ IRepository = &repositoryImpl{}

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

// Finish ...
func (r *repositoryImpl) Finish() {
	r.sess.Finish()
}

type dbRepoProviderImpl struct {
	blacklistRepo repository.Blacklist
}

var _ IRepositoryProvider = &dbRepoProviderImpl{}

// NewDBRepoProvider ...
func NewDBRepoProvider(blacklistRepo repository.Blacklist) IRepositoryProvider {
	return &dbRepoProviderImpl{
		blacklistRepo: blacklistRepo,
	}
}

// NewRepo ...
func (p *dbRepoProviderImpl) NewRepo() IRepository {
	return &dbRepoImpl{
		blacklistRepo: p.blacklistRepo,

		blacklistCustomerInputSet: map[string]struct{}{},
		blacklistCustomerOutputs:  map[string]model.BlacklistCustomer{},

		blacklistMerchantInputSet: map[string]struct{}{},
		blacklistMerchantOutputs:  map[string]model.BlacklistMerchant{},
	}
}

type dbRepoImpl struct {
	blacklistRepo repository.Blacklist

	fetchNew bool

	blacklistCustomerInputs   []string
	blacklistCustomerInputSet map[string]struct{}
	blacklistCustomerOutputs  map[string]model.BlacklistCustomer

	blacklistMerchantInputs   []string
	blacklistMerchantInputSet map[string]struct{}
	blacklistMerchantOutputs  map[string]model.BlacklistMerchant
}

var _ IRepository = &dbRepoImpl{}

func (r *dbRepoImpl) fetchData(ctx context.Context) error {
	if !r.fetchNew {
		return nil
	}
	r.fetchNew = false

	if len(r.blacklistCustomerInputs) > 0 {
		inputs := r.blacklistCustomerInputs

		keys := make([]repository.BlacklistCustomerKey, 0, len(inputs))
		for _, phone := range inputs {
			keys = append(keys, repository.BlacklistCustomerKey{
				Hash:  util.HashFunc(phone),
				Phone: phone,
			})
		}

		customers, err := r.blacklistRepo.GetBlacklistCustomers(ctx, keys)
		if err != nil {
			return err
		}

		for _, c := range customers {
			r.blacklistCustomerOutputs[c.Phone] = c
		}
	}

	if len(r.blacklistMerchantInputs) > 0 {
		inputs := r.blacklistMerchantInputs

		keys := make([]repository.BlacklistMerchantKey, 0, len(inputs))
		for _, code := range inputs {
			keys = append(keys, repository.BlacklistMerchantKey{
				Hash:         util.HashFunc(code),
				MerchantCode: code,
			})
		}

		merchants, err := r.blacklistRepo.GetBlacklistMerchants(ctx, keys)
		if err != nil {
			return err
		}
		for _, m := range merchants {
			r.blacklistMerchantOutputs[m.MerchantCode] = m
		}
	}

	return nil
}

// GetBlacklistCustomer ...
func (r *dbRepoImpl) GetBlacklistCustomer(
	ctx context.Context, phone string,
) func() (model.NullBlacklistCustomer, error) {
	r.fetchNew = true

	if _, existed := r.blacklistCustomerInputSet[phone]; !existed {
		r.blacklistCustomerInputSet[phone] = struct{}{}
		r.blacklistCustomerInputs = append(r.blacklistCustomerInputs, phone)
	}

	return func() (model.NullBlacklistCustomer, error) {
		if err := r.fetchData(ctx); err != nil {
			return model.NullBlacklistCustomer{}, err
		}

		customer, existed := r.blacklistCustomerOutputs[phone]
		if !existed {
			return model.NullBlacklistCustomer{}, nil
		}
		return model.NullBlacklistCustomer{
			Valid:    true,
			Customer: customer,
		}, nil
	}
}

// GetBlacklistMerchant ...
func (r *dbRepoImpl) GetBlacklistMerchant(
	ctx context.Context, merchantCode string,
) func() (model.NullBlacklistMerchant, error) {
	r.fetchNew = true

	if _, existed := r.blacklistMerchantInputSet[merchantCode]; !existed {
		r.blacklistMerchantInputSet[merchantCode] = struct{}{}
		r.blacklistMerchantInputs = append(r.blacklistMerchantInputs, merchantCode)
	}

	return func() (model.NullBlacklistMerchant, error) {
		if err := r.fetchData(ctx); err != nil {
			return model.NullBlacklistMerchant{}, err
		}

		merchant, existed := r.blacklistMerchantOutputs[merchantCode]
		if !existed {
			return model.NullBlacklistMerchant{}, nil
		}
		return model.NullBlacklistMerchant{
			Valid:    true,
			Merchant: merchant,
		}, nil
	}
}

// Finish ...
func (r *dbRepoImpl) Finish() {
}
