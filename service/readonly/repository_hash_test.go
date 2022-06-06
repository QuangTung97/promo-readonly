package readonly

import (
	"context"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlacklistCustomerHashDB__GetSizeLog__For_Empty_Customers(t *testing.T) {
	repo := &repository.BlacklistMock{}
	db := newBlacklistCustomerHashDB(repo)

	repo.GetConfigFunc = func(ctx context.Context) (model.BlacklistConfig, error) {
		return model.BlacklistConfig{
			CustomerCount: 0,
		}, nil
	}

	fn1 := db.GetSizeLog(newContext())
	fn2 := db.GetSizeLog(newContext())

	num, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(0), num)

	_, _ = fn2()

	assert.Equal(t, 1, len(repo.GetConfigCalls()))
}

func TestBlacklistCustomerHashDB__GetSizeLog__Multiple_Customers(t *testing.T) {
	repo := &repository.BlacklistMock{}
	db := newBlacklistCustomerHashDB(repo)

	repo.GetConfigFunc = func(ctx context.Context) (model.BlacklistConfig, error) {
		return model.BlacklistConfig{
			CustomerCount: 15,
		}, nil
	}

	fn1 := db.GetSizeLog(newContext())
	fn2 := db.GetSizeLog(newContext())

	num, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(4), num)

	_, _ = fn2()

	assert.Equal(t, 1, len(repo.GetConfigCalls()))
}

func TestBlacklistCustomerHashDB__Select_Entries(t *testing.T) {
	repo := &repository.BlacklistMock{}
	db := newBlacklistCustomerHashDB(repo)

	repo.SelectBlacklistCustomersFunc = func(
		ctx context.Context, ranges []repository.HashRange,
	) ([]model.BlacklistCustomer, error) {
		return nil, nil
	}

	fn1 := db.SelectEntries(newContext(), 20, newNullUint32(100))
	fn2 := db.SelectEntries(newContext(), 220, dhash.NullUint32{})

	_, _ = fn1()
	_, _ = fn2()

	assert.Equal(t, 1, len(repo.SelectBlacklistCustomersCalls()))
	assert.Equal(t, []repository.HashRange{
		{
			Begin: 20,
			End:   newNullUint32(100),
		},
		{
			Begin: 220,
		},
	}, repo.SelectBlacklistCustomersCalls()[0].Ranges)
}

func TestBlacklistCustomerHashDB__SelectEntries__Call_Multi_Times(t *testing.T) {
	repo := &repository.BlacklistMock{}
	db := newBlacklistCustomerHashDB(repo)

	repo.SelectBlacklistCustomersFunc = func(
		ctx context.Context, ranges []repository.HashRange,
	) ([]model.BlacklistCustomer, error) {
		return nil, nil
	}

	fn1 := db.SelectEntries(newContext(), 20, newNullUint32(100))
	_, _ = fn1()

	fn2 := db.SelectEntries(newContext(), 220, dhash.NullUint32{})
	_, _ = fn2()

	assert.Equal(t, 2, len(repo.SelectBlacklistCustomersCalls()))
	assert.Equal(t, []repository.HashRange{
		{Begin: 220},
	}, repo.SelectBlacklistCustomersCalls()[1].Ranges)
}

func TestBlacklistCustomerHashDB__Select_Entries__Returns_Correct_Data(t *testing.T) {
	repo := &repository.BlacklistMock{}
	db := newBlacklistCustomerHashDB(repo)

	repo.SelectBlacklistCustomersFunc = func(
		ctx context.Context, ranges []repository.HashRange,
	) ([]model.BlacklistCustomer, error) {
		return []model.BlacklistCustomer{
			{
				Hash:      22,
				Phone:     "0987000111",
				Status:    model.BlacklistCustomerStatusActive,
				StartTime: newNullTime("2022-05-12T10:00:00+07:00"),
				EndTime:   newNullTime("2022-05-22T10:00:00+07:00"),
			},
			{
				Hash:   300,
				Phone:  "0987000222",
				Status: model.BlacklistCustomerStatusInactive,
			},
		}, nil
	}

	fn1 := db.SelectEntries(newContext(), 20, newNullUint32(100))
	fn2 := db.SelectEntries(newContext(), 220, dhash.NullUint32{})

	entries1, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, []dhash.Entry{
		{
			Hash: 22,
			Data: marshalBlacklistCustomer(model.BlacklistCustomer{
				Hash:      22,
				Phone:     "0987000111",
				Status:    model.BlacklistCustomerStatusActive,
				StartTime: newNullTime("2022-05-12T10:00:00+07:00"),
				EndTime:   newNullTime("2022-05-22T10:00:00+07:00"),
			}),
		},
	}, entries1)

	entries2, err := fn2()
	assert.Equal(t, nil, err)
	assert.Equal(t, []dhash.Entry{
		{
			Hash: 300,
			Data: marshalBlacklistCustomer(model.BlacklistCustomer{
				Hash:   300,
				Phone:  "0987000222",
				Status: model.BlacklistCustomerStatusInactive,
			}),
		},
	}, entries2)
}
