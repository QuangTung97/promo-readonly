package readonly

import (
    "context"
    "fmt"
    "github.com/QuangTung97/promo-readonly/model"
    "github.com/QuangTung97/promo-readonly/pkg/dhash"
    "github.com/QuangTung97/promo-readonly/repository"
    "github.com/spaolacci/murmur3"
    "math/bits"
)

// IRepository ...
type IRepository interface {
    GetBlacklistCustomer(ctx context.Context, phone string) func() (model.NullBlacklistCustomer, error)
    GetCampaigns(ctx context.Context, voucherCode string) func() ([]model.Campaign, error)
}

type repositoryImpl struct {
    sess dhash.Session
}

type blacklistCustomerHashDB struct {
    repo repository.Blacklist

    fetchNew bool

    fetchSizeLog bool
    sizeLog      uint64

    selectInputs   []repository.HashRange
    selectInputSet map[repository.HashRange]struct{}
    output         map[repository.HashRange][]dhash.Entry

    err error
}

func newBlacklistCustomerHashDB(repo repository.Blacklist) *blacklistCustomerHashDB {
    return &blacklistCustomerHashDB{
        repo: repo,
    }
}

func log2Int(n int64) uint64 {
    if n == 0 {
        return 0
    }
    return 64 - uint64(bits.LeadingZeros64(uint64(n-1)))
}

func (h *blacklistCustomerHashDB) fetchData(ctx context.Context) error {
    if !h.fetchNew {
        return h.err
    }
    h.fetchNew = false

    if h.fetchSizeLog {
        h.fetchSizeLog = false
        config, err := h.repo.GetConfig(ctx)
        if err != nil {
            h.err = err
            return err
        }

        h.sizeLog = log2Int(config.CustomerCount)
    }

    return nil
}

func (h *blacklistCustomerHashDB) GetSizeLog(ctx context.Context) func() (uint64, error) {
    h.fetchNew = true
    h.fetchSizeLog = true
    return func() (uint64, error) {
        if err := h.fetchData(ctx); err != nil {
            return 0, err
        }
        return h.sizeLog, nil
    }
}

func (h *blacklistCustomerHashDB) SelectEntries(
    ctx context.Context, hashBegin uint32, hashEnd dhash.NullUint32,
) func() ([]dhash.Entry, error) {
    return func() ([]dhash.Entry, error) {
        return nil, nil
    }
}

func hashFunc(s string) uint32 {
    return murmur3.Sum32([]byte(s))
}

// GetBlacklistCustomer ...
func (r *repositoryImpl) GetBlacklistCustomer(
    ctx context.Context, phone string,
) func() (model.NullBlacklistCustomer, error) {
    hash := r.sess.NewHash("bl:cst", &blacklistCustomerHashDB{})
    fn := hash.SelectEntries(ctx, hashFunc(phone))
    return func() (model.NullBlacklistCustomer, error) {
        entries, err := fn()
        if err != nil {
            return model.NullBlacklistCustomer{}, err
        }
        for _, entry := range entries {
            fmt.Println(entry.Data)
        }
        return model.NullBlacklistCustomer{}, nil
    }
}
