package repository

import (
	"context"
	"database/sql"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/integration"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func newContext() context.Context {
	return context.Background()
}

type blacklistTest struct {
	tc       *integration.TestCase
	provider Provider
}

func newBlacklistTest() *blacklistTest {
	tc := integration.NewTestCase()
	tc.Truncate("blacklist_customer")
	return &blacklistTest{
		tc:       tc,
		provider: NewProvider(tc.DB),
	}
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

func TestBlacklist(t *testing.T) {
	tc := newBlacklistTest()

	repo := NewBlacklist()

	//---------------------------------------
	// Get Customers
	//---------------------------------------
	readCtx := tc.provider.Readonly(newContext())
	customers, err := repo.GetBlacklistCustomers(readCtx, []BlacklistCustomerKey{
		{
			Hash:  3300,
			Phone: "0987000111",
		},
	})
	assert.Equal(t, nil, err)
	assert.Nil(t, customers)

	//---------------------------------------
	// Insert
	//---------------------------------------
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistCustomers(ctx, []model.BlacklistCustomer{
			{
				Hash:      3300,
				Phone:     "0987000111",
				Status:    model.BlacklistCustomerStatusActive,
				StartTime: newNullTime("2022-05-10T10:00:00+07:00"),
				EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
			},
			{
				Hash:   4400,
				Phone:  "0987000222",
				Status: model.BlacklistCustomerStatusInactive,
			},
		})
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Customers 2
	//---------------------------------------
	customers, err = repo.GetBlacklistCustomers(readCtx, []BlacklistCustomerKey{
		{
			Hash:  3300,
			Phone: "0987000111",
		},
		{
			Hash:  4400,
			Phone: "0987000222",
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, []model.BlacklistCustomer{
		{
			Hash:      3300,
			Phone:     "0987000111",
			Status:    model.BlacklistCustomerStatusActive,
			StartTime: newNullTime("2022-05-10T10:00:00+07:00"),
			EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
		},
		{
			Hash:   4400,
			Phone:  "0987000222",
			Status: model.BlacklistCustomerStatusInactive,
		},
	}, customers)

	//---------------------------------------
	// Upsert
	//---------------------------------------
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistCustomers(ctx, []model.BlacklistCustomer{
			{
				Hash:      3300,
				Phone:     "0987000111",
				Status:    model.BlacklistCustomerStatusInactive,
				StartTime: newNullTime("2022-07-10T10:00:00+07:00"),
				EndTime:   newNullTime("2022-07-18T10:00:00+07:00"),
			},
			{
				Hash:   4400,
				Phone:  "0987000222",
				Status: model.BlacklistCustomerStatusActive,
			},
		})
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Customers 3
	//---------------------------------------
	customers, err = repo.GetBlacklistCustomers(readCtx, []BlacklistCustomerKey{
		{
			Hash:  3300,
			Phone: "0987000111",
		},
		{
			Hash:  4400,
			Phone: "0987000222",
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, []model.BlacklistCustomer{
		{
			Hash:      3300,
			Phone:     "0987000111",
			Status:    model.BlacklistCustomerStatusInactive,
			StartTime: newNullTime("2022-07-10T10:00:00+07:00"),
			EndTime:   newNullTime("2022-07-18T10:00:00+07:00"),
		},
		{
			Hash:   4400,
			Phone:  "0987000222",
			Status: model.BlacklistCustomerStatusActive,
		},
	}, customers)
}
