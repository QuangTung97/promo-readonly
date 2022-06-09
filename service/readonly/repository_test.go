package readonly

import (
	"context"
	"database/sql"
	"errors"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/pkg/util"
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
	blacklistMerchantHash *dhash.HashMock

	repo IRepository
}

func newRepoTest() *repoTest {
	sess := &dhash.SessionMock{}
	blacklistMerchantHash := &dhash.HashMock{}
	return &repoTest{
		blacklistMerchantHash: blacklistMerchantHash,

		repo: newRepository(sess, nil, blacklistMerchantHash),
	}
}

func (r *repoTest) stubMerchantSelectEntries(entries []dhash.Entry, err error) {
	r.blacklistMerchantHash.SelectEntriesFunc = func(ctx context.Context, hash uint32) func() ([]dhash.Entry, error) {
		return func() ([]dhash.Entry, error) {
			return entries, err
		}
	}
}

func TestRepository_GetBlacklistMerchant__Call_Correct_Select_Entries(t *testing.T) {
	r := newRepoTest()

	r.stubMerchantSelectEntries(nil, nil)

	r.repo.GetBlacklistMerchant(newContext(), "MERCHANT01")

	assert.Equal(t, 1, len(r.blacklistMerchantHash.SelectEntriesCalls()))
	assert.Equal(t, util.HashFunc("MERCHANT01"), r.blacklistMerchantHash.SelectEntriesCalls()[0].Hash)
}

func TestRepository_GetBlacklistMerchant__Select_Entries__Returns_Error(t *testing.T) {
	r := newRepoTest()

	someErr := errors.New("some error")
	r.stubMerchantSelectEntries(nil, someErr)

	fn := r.repo.GetBlacklistMerchant(newContext(), "MERCHANT01")
	merchant, err := fn()
	assert.Equal(t, someErr, err)
	assert.Equal(t, model.NullBlacklistMerchant{}, merchant)
}

func TestRepository_GetBlacklistMerchant__Select_Entries__Returns_Empty(t *testing.T) {
	r := newRepoTest()

	r.stubMerchantSelectEntries(nil, nil)

	fn := r.repo.GetBlacklistMerchant(newContext(), "MERCHANT01")
	merchant, err := fn()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{}, merchant)
}

func TestRepository_GetBlacklistMerchant__Select_Entries__Returns_OK(t *testing.T) {
	r := newRepoTest()

	merchantCode := "MERCHANT01"
	hash := util.HashFunc(merchantCode)

	merchant := model.BlacklistMerchant{
		Hash:         hash,
		MerchantCode: merchantCode,
		Status:       model.BlacklistMerchantStatusActive,
	}

	r.stubMerchantSelectEntries([]dhash.Entry{
		{
			Hash: hash,
			Data: marshalBlacklistMerchant(merchant),
		},
	}, nil)

	fn := r.repo.GetBlacklistMerchant(newContext(), merchantCode)
	nullMerchant, err := fn()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{
		Valid:    true,
		Merchant: merchant,
	}, nullMerchant)
}

func TestRepository_GetBlacklistMerchant__Select_Entries__Hash_Mismatch(t *testing.T) {
	r := newRepoTest()

	merchantCode := "MERCHANT01"
	hash := util.HashFunc(merchantCode)

	merchant := model.BlacklistMerchant{
		Hash:         hash,
		MerchantCode: merchantCode,
		Status:       model.BlacklistMerchantStatusActive,
	}

	r.stubMerchantSelectEntries([]dhash.Entry{
		{
			Hash: hash + 1,
			Data: marshalBlacklistMerchant(merchant),
		},
	}, nil)

	fn := r.repo.GetBlacklistMerchant(newContext(), merchantCode)
	nullMerchant, err := fn()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{}, nullMerchant)
}

func TestRepository_GetBlacklistMerchant__Select_Entries__Code_Mismatch(t *testing.T) {
	r := newRepoTest()

	merchantCode := "MERCHANT01"
	hash := util.HashFunc(merchantCode)

	merchant := model.BlacklistMerchant{
		Hash:         hash,
		MerchantCode: "MERCHANT02",
		Status:       model.BlacklistMerchantStatusActive,
	}

	r.stubMerchantSelectEntries([]dhash.Entry{
		{
			Hash: hash,
			Data: marshalBlacklistMerchant(merchant),
		},
	}, nil)

	fn := r.repo.GetBlacklistMerchant(newContext(), merchantCode)
	nullMerchant, err := fn()
	assert.Equal(t, nil, err)
	assert.Equal(t, model.NullBlacklistMerchant{}, nullMerchant)
}
