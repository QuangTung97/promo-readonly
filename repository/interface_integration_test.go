package repository

import (
	"context"
	"github.com/QuangTung97/promo-readonly/pkg/integration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider_Readonly__GetReadonly(t *testing.T) {
	tc := integration.NewTestCase()

	p := NewProvider(tc.DB)
	ctx := p.Readonly(newContext())

	db := GetReadonly(ctx)

	var version string
	err := db.GetContext(ctx, &version, "SELECT VERSION()")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", version)
}

func TestProvider_Transact__GetTransaction(t *testing.T) {
	tc := integration.NewTestCase()

	var version string

	p := NewProvider(tc.DB)
	err := p.Transact(newContext(), func(ctx context.Context) error {
		tx := GetTx(ctx)

		err := tx.GetContext(ctx, &version, "SELECT VERSION()")
		assert.Equal(t, nil, err)

		return nil
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", version)
}

func TestProvider_Transact__GetReadonly(t *testing.T) {
	tc := integration.NewTestCase()

	var version string

	p := NewProvider(tc.DB)
	err := p.Transact(newContext(), func(ctx context.Context) error {
		db := GetReadonly(ctx)

		err := db.GetContext(ctx, &version, "SELECT VERSION()")
		assert.Equal(t, nil, err)

		return nil
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", version)
}

func TestProvider_Transact__Multi_Calls_Multi_Levels(t *testing.T) {
	tc := integration.NewTestCase()

	var version string

	p := NewProvider(tc.DB)
	err := p.Transact(newContext(), func(ctx context.Context) error {
		return p.Transact(ctx, func(ctx context.Context) error {
			tx := GetTx(ctx)

			err := tx.GetContext(ctx, &version, "SELECT VERSION()")
			assert.Equal(t, nil, err)

			return nil
		})
	})
	assert.Equal(t, nil, err)
	assert.NotEqual(t, "", version)
}
