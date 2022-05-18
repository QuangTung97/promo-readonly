package readonly

import (
    "context"
    "github.com/QuangTung97/promo-readonly/model"
    "github.com/QuangTung97/promo-readonly/pkg/dhash"
    "github.com/QuangTung97/promo-readonly/repository"
    "github.com/stretchr/testify/assert"
    "testing"
)

func newContext() context.Context {
    return context.Background()
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

    fn1 := db.SelectEntries(newContext(), 20, newNullUint32(100))
    fn2 := db.SelectEntries(newContext(), 220, dhash.NullUint32{})

    entries, err := fn1()
    assert.Equal(t, nil, err)
    assert.Equal(t, nil, entries)

    entries, err = fn2()
    assert.Equal(t, nil, err)
    assert.Equal(t, nil, entries)
}
