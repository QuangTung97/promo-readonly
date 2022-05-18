package readonly

import (
	"context"
	"database/sql"
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

func newContext() context.Context {
	return context.Background()
}

func newTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

func newNullTime(s string) sql.NullTime {
	return sql.NullTime{
		Valid: true,
		Time:  newTime(s),
	}
}

func newNullUint32(v uint32) dhash.NullUint32 {
	return dhash.NullUint32{
		Valid: true,
		Num:   v,
	}
}

func TestLog2Int(t *testing.T) {
	assert.Equal(t, uint64(0), log2Int(0))
	assert.Equal(t, uint64(0), log2Int(1))
	assert.Equal(t, uint64(2), log2Int(4))

	assert.Equal(t, uint64(4), log2Int(15))
	assert.Equal(t, uint64(4), log2Int(16))

	assert.Equal(t, uint64(6), log2Int(62))
	assert.Equal(t, uint64(6), log2Int(63))
	assert.Equal(t, uint64(6), log2Int(64))
	assert.Equal(t, uint64(7), log2Int(65))
}

type repoTest struct {
	client    *cacheclient.Client
	provider  repository.Provider
	mem       *memtable.MemTable
	repo      IRepository
	blacklist repository.Blacklist
}

func newRepoTest(tc *integration.TestCase) *repoTest {
	tc.Truncate("blacklist_config")
	tc.Truncate("blacklist_customer")

	client := cacheclient.New("localhost:11211", 1)
	err := client.UnsafeFlushAll()
	if err != nil {
		panic(err)
	}

	repoProvider := repository.NewProvider(tc.DB)

	blacklistRepo := repository.NewBlacklist()

	mem := memtable.New(100 * 1024)

	provider := dhash.NewProvider(mem, client)
	sess := provider.NewSession()
	repo := NewRepository(sess, blacklistRepo)

	return &repoTest{
		client:    client,
		provider:  repoProvider,
		mem:       mem,
		repo:      repo,
		blacklist: blacklistRepo,
	}
}

func (r *repoTest) finish() {
	err := r.client.Close()
	if err != nil {
		panic(err)
	}
}

func TestRepository_GetBlacklistCustomer__Found(t *testing.T) {
	tc := integration.NewTestCase()
	r := newRepoTest(tc)
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
