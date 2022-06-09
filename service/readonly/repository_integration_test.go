// +build integration

package readonly

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/cacheclient"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/integration"
	"github.com/QuangTung97/promo-readonly/pkg/memtable"
	"github.com/QuangTung97/promo-readonly/pkg/util"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type repoIntegrationTest struct {
	client    *cacheclient.Client
	provider  repository.Provider
	mem       *memtable.MemTable
	repo      IRepository
	blacklist repository.Blacklist
}

func newRepoIntegrationTest(tc *integration.TestCase) *repoIntegrationTest {
	tc.Truncate("blacklist_config")
	tc.Truncate("blacklist_customer")

	client := cacheclient.New("localhost:11211", 1)
	err := client.UnsafeFlushAll()
	if err != nil {
		panic(err)
	}

	txProvider := repository.NewProvider(tc.DB)

	blacklistRepo := repository.NewBlacklist()

	mem := memtable.New(100 * 1024)

	dhashProvider := dhash.NewProvider(mem, client)
	repoProvider := NewRepositoryProvider(dhashProvider, blacklistRepo)
	repo := repoProvider.NewRepo()

	return &repoIntegrationTest{
		client:    client,
		provider:  txProvider,
		mem:       mem,
		repo:      repo,
		blacklist: blacklistRepo,
	}
}

func (r *repoIntegrationTest) finish() {
	err := r.client.Close()
	if err != nil {
		panic(err)
	}
}

func TestRepository_GetBlacklistCustomer(t *testing.T) {
	tc := integration.NewTestCase()
	r := newRepoIntegrationTest(tc)
	defer r.finish()

	err := r.provider.Transact(newContext(), func(ctx context.Context) error {
		return r.blacklist.UpsertBlacklistCustomers(ctx, []model.BlacklistCustomer{
			{
				Hash:      util.HashFunc("0987000111"),
				Phone:     "0987000111",
				Status:    model.BlacklistCustomerStatusActive,
				StartTime: newNullTime("2022-05-08T10:00:00+07:00"),
				EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
			},
		})
	})
	assert.Equal(t, nil, err)

	ctx := r.provider.Readonly(newContext())

	start := time.Now()
	fn1 := r.repo.GetBlacklistCustomer(ctx, "0987000111")
	fn2 := r.repo.GetBlacklistCustomer(ctx, "0987000222")

	customer1, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistCustomer{
		Valid: true,
		Customer: model.BlacklistCustomer{
			Hash:      util.HashFunc("0987000111"),
			Phone:     "0987000111",
			Status:    model.BlacklistCustomerStatusActive,
			StartTime: newNullTime("2022-05-08T10:00:00+07:00"),
			EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
		},
	}, customer1)

	customer2, err := fn2()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistCustomer{}, customer2)

	fmt.Println("First Get:", time.Since(start))

	// Get Second Times
	start = time.Now()
	fn1 = r.repo.GetBlacklistCustomer(ctx, "0987000111")
	customer1, err = fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistCustomer{
		Valid: true,
		Customer: model.BlacklistCustomer{
			Hash:      util.HashFunc("0987000111"),
			Phone:     "0987000111",
			Status:    model.BlacklistCustomerStatusActive,
			StartTime: newNullTime("2022-05-08T10:00:00+07:00"),
			EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
		},
	}, customer1)
	fmt.Println("Second Get:", time.Since(start))

	// Get Mem
	num, ok := r.mem.GetNum("bl:cst")
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(0), num)

	// Get Cache
	pipe := r.client.Pipeline()
	getOutput, err := pipe.Get("bl:cst:size-log")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.GetOutput{
		Found: true, Data: []byte("0"),
	}, getOutput)

	getOutput, err = pipe.Get("bl:cst:0:00000000")()
	assert.Equal(t, nil, err)
	assert.Equal(t, true, getOutput.Found)

	// Get Third Times
	start = time.Now()
	fn1 = r.repo.GetBlacklistCustomer(ctx, "0987000111")
	customer1, err = fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistCustomer{
		Valid: true,
		Customer: model.BlacklistCustomer{
			Hash:      util.HashFunc("0987000111"),
			Phone:     "0987000111",
			Status:    model.BlacklistCustomerStatusActive,
			StartTime: newNullTime("2022-05-08T10:00:00+07:00"),
			EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
		},
	}, customer1)
	fmt.Println("Third Get:", time.Since(start))
}

func TestRepository_GetBlacklistMerchant__Integration(t *testing.T) {
	tc := integration.NewTestCase()
	r := newRepoIntegrationTest(tc)
	defer r.finish()

	err := r.provider.Transact(newContext(), func(ctx context.Context) error {
		return r.blacklist.UpsertBlacklistMerchants(ctx, []model.BlacklistMerchant{
			{
				Hash:         util.HashFunc("MERCHANT01"),
				MerchantCode: "MERCHANT01",
				Status:       model.BlacklistMerchantStatusActive,
				StartTime:    newNullTime("2022-05-08T10:00:00+07:00"),
				EndTime:      newNullTime("2022-05-18T10:00:00+07:00"),
			},
		})
	})
	assert.Equal(t, nil, err)

	ctx := r.provider.Readonly(newContext())

	fn1 := r.repo.GetBlacklistMerchant(ctx, "MERCHANT01")
	fn2 := r.repo.GetBlacklistMerchant(ctx, "MERCHANT02")

	start := time.Now()

	merchant1, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{
		Valid: true,
		Merchant: model.BlacklistMerchant{
			Hash:         util.HashFunc("MERCHANT01"),
			MerchantCode: "MERCHANT01",
			Status:       model.BlacklistMerchantStatusActive,
			StartTime:    newNullTime("2022-05-08T10:00:00+07:00"),
			EndTime:      newNullTime("2022-05-18T10:00:00+07:00"),
		},
	}, merchant1)

	merchant2, err := fn2()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{}, merchant2)

	fmt.Println("First Get:", time.Since(start))

	// Get Second Times
	start = time.Now()
	fn1 = r.repo.GetBlacklistMerchant(ctx, "MERCHANT01")
	merchant1, err = fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, true, merchant1.Valid)
	fmt.Println("Second Get:", time.Since(start))

	// Get Mem
	num, ok := r.mem.GetNum("bl:mc")
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(0), num)

	// Get Cache
	pipe := r.client.Pipeline()
	getOutput, err := pipe.Get("bl:mc:size-log")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.GetOutput{
		Found: true, Data: []byte("0"),
	}, getOutput)

	getOutput, err = pipe.Get("bl:mc:0:00000000")()
	assert.Equal(t, nil, err)
	assert.Equal(t, true, getOutput.Found)

	// Get Third Times
	start = time.Now()
	fn1 = r.repo.GetBlacklistMerchant(ctx, "MERCHANT01")
	merchant1, err = fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, true, merchant1.Valid)
	fmt.Println("Third Get:", time.Since(start))
}
