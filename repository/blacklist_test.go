package repository

import (
	"context"
	"database/sql"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
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

func newNullUint32(v uint32) dhash.NullUint32 {
	return dhash.NullUint32{
		Valid: true,
		Num:   v,
	}
}

func TestBlacklist_Empty_Keys(t *testing.T) {
	repo := NewBlacklist()
	ctx := newContext()

	customers, err := repo.GetBlacklistCustomers(ctx, nil)
	assert.Equal(t, nil, err)
	assert.Nil(t, customers)

	customers, err = repo.SelectBlacklistCustomers(ctx, nil)
	assert.Equal(t, nil, err)
	assert.Nil(t, customers)

	merchants, err := repo.GetBlacklistMerchants(ctx, nil)
	assert.Equal(t, nil, err)
	assert.Nil(t, merchants)

	terminals, err := repo.GetBlacklistTerminals(ctx, nil)
	assert.Equal(t, nil, err)
	assert.Nil(t, terminals)

	err = repo.UpsertBlacklistCustomers(ctx, nil)
	assert.Equal(t, nil, err)

	err = repo.UpsertBlacklistMerchants(ctx, nil)
	assert.Equal(t, nil, err)

	err = repo.UpsertBlacklistTerminals(ctx, nil)
	assert.Equal(t, nil, err)

	merchants, err = repo.SelectBlacklistMerchants(ctx, nil)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(merchants))
}

func TestBlacklist_Customers(t *testing.T) {
	tc := newBlacklistTest()
	tc.tc.Truncate("blacklist_customer")

	repo := NewBlacklist()

	const hash01 = 3300
	const phone01 = "0987000111"
	const hash02 = 4400
	const phone02 = "0987000222"

	key01 := BlacklistCustomerKey{Hash: hash01, Phone: phone01}
	key02 := BlacklistCustomerKey{Hash: hash02, Phone: phone02}

	//---------------------------------------
	// Get Customers
	//---------------------------------------
	readCtx := tc.provider.Readonly(newContext())
	customers, err := repo.GetBlacklistCustomers(readCtx, []BlacklistCustomerKey{key01})
	assert.Equal(t, nil, err)
	assert.Nil(t, customers)

	//---------------------------------------
	// Insert
	//---------------------------------------
	insertCustomers := []model.BlacklistCustomer{
		{
			Hash:      hash01,
			Phone:     phone01,
			Status:    model.BlacklistCustomerStatusActive,
			StartTime: newNullTime("2022-05-10T10:00:00+07:00"),
			EndTime:   newNullTime("2022-05-18T10:00:00+07:00"),
		},
		{
			Hash:   hash02,
			Phone:  phone02,
			Status: model.BlacklistCustomerStatusInactive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistCustomers(ctx, insertCustomers)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Customers 2
	//---------------------------------------
	keys := []BlacklistCustomerKey{key01, key02}

	customers, err = repo.GetBlacklistCustomers(readCtx, keys)
	assert.Equal(t, nil, err)
	assert.Equal(t, insertCustomers, customers)

	//---------------------------------------
	// Upsert
	//---------------------------------------
	upsertCustomers := []model.BlacklistCustomer{
		{
			Hash:      hash01,
			Phone:     phone01,
			Status:    model.BlacklistCustomerStatusInactive,
			StartTime: newNullTime("2022-07-10T10:00:00+07:00"),
			EndTime:   newNullTime("2022-07-18T10:00:00+07:00"),
		},
		{
			Hash:   hash02,
			Phone:  phone02,
			Status: model.BlacklistCustomerStatusActive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistCustomers(ctx, upsertCustomers)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Customers 3
	//---------------------------------------
	customers, err = repo.GetBlacklistCustomers(readCtx, keys)
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertCustomers, customers)

	//---------------------------------------
	// Select Hash Range
	//---------------------------------------
	customers, err = repo.SelectBlacklistCustomers(readCtx, []HashRange{
		{
			Begin: 3300,
			End:   newNullUint32(4401),
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertCustomers, customers)

	//---------------------------------------
	// Select Hash Range 2
	//---------------------------------------
	customers, err = repo.SelectBlacklistCustomers(readCtx, []HashRange{
		{
			Begin: 3300,
			End:   newNullUint32(3301),
		},
		{
			Begin: 4400,
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertCustomers, customers)
}

func TestBlacklist_Merchants(t *testing.T) {
	tc := newBlacklistTest()
	tc.tc.Truncate("blacklist_merchant")

	repo := NewBlacklist()

	const hash01 = 3300
	const merchantCode01 = "MERCHANT01"
	const hash02 = 4400
	const merchantCode02 = "MERCHANT02"

	ctx := tc.provider.Readonly(newContext())

	key01 := BlacklistMerchantKey{Hash: hash01, MerchantCode: merchantCode01}
	key02 := BlacklistMerchantKey{Hash: hash02, MerchantCode: merchantCode02}

	//---------------------------------------
	// Get Merchants 1
	//---------------------------------------
	merchants, err := repo.GetBlacklistMerchants(ctx, []BlacklistMerchantKey{key01})
	assert.Equal(t, nil, err)
	assert.Nil(t, merchants)

	//---------------------------------------
	// Insert Merchants
	//---------------------------------------
	insertMerchants := []model.BlacklistMerchant{
		{
			Hash:         hash01,
			MerchantCode: merchantCode01,
			Status:       model.BlacklistMerchantStatusActive,
			StartTime:    newNullTime("2022-05-11T10:00:00+07:00"),
			EndTime:      newNullTime("2022-05-18T10:00:00+07:00"),
		},
		{
			Hash:         hash02,
			MerchantCode: merchantCode02,
			Status:       model.BlacklistMerchantStatusInactive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistMerchants(ctx, insertMerchants)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Merchants 2
	//---------------------------------------
	merchants, err = repo.GetBlacklistMerchants(ctx, []BlacklistMerchantKey{key01, key02})
	assert.Equal(t, nil, err)
	assert.Equal(t, insertMerchants, merchants)

	//---------------------------------------
	// Upsert Merchants
	//---------------------------------------
	upsertMerchants := []model.BlacklistMerchant{
		{
			Hash:         hash01,
			MerchantCode: merchantCode01,
			Status:       model.BlacklistMerchantStatusInactive,
			StartTime:    newNullTime("2022-07-11T10:00:00+07:00"),
			EndTime:      newNullTime("2022-07-18T10:00:00+07:00"),
		},
		{
			Hash:         hash02,
			MerchantCode: merchantCode02,
			Status:       model.BlacklistMerchantStatusActive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistMerchants(ctx, upsertMerchants)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Merchants 3
	//---------------------------------------
	merchants, err = repo.GetBlacklistMerchants(ctx, []BlacklistMerchantKey{key01, key02})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertMerchants, merchants)

	//---------------------------------------
	// Select Merchants
	//---------------------------------------
	merchants, err = repo.SelectBlacklistMerchants(ctx, []HashRange{
		{
			Begin: hash01,
			End:   newNullUint32(hash01 + 1),
		},
		{
			Begin: hash02,
			End:   newNullUint32(hash02 + 1),
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertMerchants, merchants)

	//---------------------------------------
	// Select Merchants 1
	//---------------------------------------
	merchants, err = repo.SelectBlacklistMerchants(ctx, []HashRange{
		{
			Begin: hash01,
			End:   newNullUint32(hash01 + 1),
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertMerchants[:1], merchants)

	//---------------------------------------
	// Select Merchants 2
	//---------------------------------------
	merchants, err = repo.SelectBlacklistMerchants(ctx, []HashRange{
		{
			Begin: hash01,
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertMerchants, merchants)
}

func TestBlacklist_Terminals(t *testing.T) {
	tc := newBlacklistTest()
	tc.tc.Truncate("blacklist_terminal")

	repo := NewBlacklist()

	const hash01 = 3300
	const merchantCode01 = "MERCHANT01"
	const terminalCode01 = "TERM01"

	const hash02 = 4400
	const merchantCode02 = "MERCHANT02"
	const terminalCode02 = "TERM02"

	ctx := tc.provider.Readonly(newContext())

	key01 := BlacklistTerminalKey{Hash: hash01, MerchantCode: merchantCode01, TerminalCode: terminalCode01}
	key02 := BlacklistTerminalKey{Hash: hash02, MerchantCode: merchantCode02, TerminalCode: terminalCode02}

	//---------------------------------------
	// Get Terminals 1
	//---------------------------------------
	terminals, err := repo.GetBlacklistTerminals(ctx, []BlacklistTerminalKey{key01})
	assert.Equal(t, nil, err)
	assert.Nil(t, terminals)

	//---------------------------------------
	// Insert Terminals
	//---------------------------------------
	insertTerminals := []model.BlacklistTerminal{
		{
			Hash:         hash01,
			MerchantCode: merchantCode01,
			TerminalCode: terminalCode01,
			Status:       model.BlacklistTerminalStatusActive,
			StartTime:    newNullTime("2022-05-11T10:00:00+07:00"),
			EndTime:      newNullTime("2022-05-18T10:00:00+07:00"),
		},
		{
			Hash:         hash02,
			MerchantCode: merchantCode02,
			TerminalCode: terminalCode02,
			Status:       model.BlacklistTerminalStatusInactive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistTerminals(ctx, insertTerminals)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Terminals 2
	//---------------------------------------
	terminals, err = repo.GetBlacklistTerminals(ctx, []BlacklistTerminalKey{key01, key02})
	assert.Equal(t, nil, err)
	assert.Equal(t, insertTerminals, terminals)

	//---------------------------------------
	// Upsert Terminals
	//---------------------------------------
	upsertTerminals := []model.BlacklistTerminal{
		{
			Hash:         hash01,
			MerchantCode: merchantCode01,
			TerminalCode: terminalCode01,
			Status:       model.BlacklistTerminalStatusInactive,
			StartTime:    newNullTime("2022-07-11T10:00:00+07:00"),
			EndTime:      newNullTime("2022-07-18T10:00:00+07:00"),
		},
		{
			Hash:         hash02,
			MerchantCode: merchantCode02,
			TerminalCode: terminalCode02,
			Status:       model.BlacklistTerminalStatusActive,
		},
	}
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertBlacklistTerminals(ctx, upsertTerminals)
	})
	assert.Equal(t, nil, err)

	//---------------------------------------
	// Get Terminals 3
	//---------------------------------------
	terminals, err = repo.GetBlacklistTerminals(ctx, []BlacklistTerminalKey{key01, key02})
	assert.Equal(t, nil, err)
	assert.Equal(t, upsertTerminals, terminals)
}

func TestBlacklist_Config(t *testing.T) {
	tc := newBlacklistTest()
	tc.tc.Truncate("blacklist_config")

	repo := NewBlacklist()

	ctx := tc.provider.Readonly(newContext())

	// Get Config
	config, err := repo.GetConfig(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, model.BlacklistConfig{}, config)

	// Insert
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertConfig(ctx, model.BlacklistConfig{
			CustomerCount: 10,
			MerchantCount: 11,
			TerminalCount: 12,
		})
	})
	assert.Equal(t, nil, err)

	// Get Config
	config, err = repo.GetConfig(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, model.BlacklistConfig{
		ID:            1,
		CustomerCount: 10,
		MerchantCount: 11,
		TerminalCount: 12,
	}, config)

	// Upsert
	err = tc.provider.Transact(newContext(), func(ctx context.Context) error {
		return repo.UpsertConfig(ctx, model.BlacklistConfig{
			ID:            1,
			CustomerCount: 20,
			MerchantCount: 21,
			TerminalCount: 22,
		})
	})
	assert.Equal(t, nil, err)

	// Get
	config, err = repo.GetConfig(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, model.BlacklistConfig{
		ID:            1,
		CustomerCount: 20,
		MerchantCount: 21,
		TerminalCount: 22,
	}, config)
}
