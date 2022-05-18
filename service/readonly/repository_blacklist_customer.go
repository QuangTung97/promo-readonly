package readonly

import (
	"context"
	"database/sql"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/promopb"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
)

func newTimestampNull(t sql.NullTime) *timestamp.Timestamp {
	if !t.Valid {
		return nil
	}
	return timestamppb.New(t.Time)
}

func nullTimeFromTimestamp(t *timestamp.Timestamp) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Valid: true,
		Time:  t.AsTime(),
	}
}

func marshalBlacklistCustomer(c model.BlacklistCustomer) []byte {
	msg := promopb.BlacklistCustomerData{
		Hash:      c.Hash,
		Phone:     c.Phone,
		Status:    uint32(c.Status),
		StartTime: newTimestampNull(c.StartTime),
		EndTime:   newTimestampNull(c.EndTime),
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return data
}

func unmarshalBlacklistCustomer(data []byte) (model.BlacklistCustomer, error) {
	var msg promopb.BlacklistCustomerData
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return model.BlacklistCustomer{}, err
	}
	return model.BlacklistCustomer{
		Hash:      msg.Hash,
		Phone:     msg.Phone,
		Status:    model.BlacklistCustomerStatus(msg.Status),
		StartTime: nullTimeFromTimestamp(msg.StartTime),
		EndTime:   nullTimeFromTimestamp(msg.EndTime),
	}, nil
}

type blacklistCustomerHashDB struct {
	repo repository.Blacklist

	fetchNew bool

	fetchSizeLog bool
	sizeLog      uint64

	selectInputs   []repository.HashRange
	selectInputSet map[repository.HashRange]struct{}
	customers      []model.BlacklistCustomer

	err error
}

func newBlacklistCustomerHashDB(repo repository.Blacklist) *blacklistCustomerHashDB {
	return &blacklistCustomerHashDB{
		repo: repo,
	}
}

func (h *blacklistCustomerHashDB) fetchData(ctx context.Context) error {
	if h.err != nil {
		return h.err
	}
	h.err = h.fetchDataWithError(ctx)
	return h.err
}

func (h *blacklistCustomerHashDB) fetchDataWithError(ctx context.Context) error {
	if !h.fetchNew {
		return h.err
	}
	h.fetchNew = false

	if h.fetchSizeLog {
		h.fetchSizeLog = false
		config, err := h.repo.GetConfig(ctx)
		if err != nil {
			return err
		}

		h.sizeLog = log2Int(config.CustomerCount)
	}

	if len(h.selectInputs) > 0 {
		inputs := h.selectInputs
		h.selectInputs = nil

		customers, err := h.repo.SelectBlacklistCustomers(ctx, inputs)
		if err != nil {
			return err
		}
		sort.Slice(customers, func(i, j int) bool {
			return customers[i].Hash < customers[j].Hash
		})
		h.customers = customers
	}

	return nil
}

// GetSizeLog ...
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

// SelectEntries ...
func (h *blacklistCustomerHashDB) SelectEntries(
	ctx context.Context, hashBegin uint32, hashEnd dhash.NullUint32,
) func() ([]dhash.Entry, error) {
	h.fetchNew = true

	h.selectInputs = append(h.selectInputs, repository.HashRange{
		Begin: hashBegin,
		End:   hashEnd,
	})

	return func() ([]dhash.Entry, error) {
		if err := h.fetchData(ctx); err != nil {
			return nil, err
		}

		var result []dhash.Entry
		for _, customer := range h.customers {
			if customer.Hash < hashBegin {
				continue
			}
			if hashEnd.Valid && customer.Hash >= hashEnd.Num {
				continue
			}
			result = append(result, dhash.Entry{
				Hash: customer.Hash,
				Data: marshalBlacklistCustomer(customer),
			})
		}
		return result, nil
	}
}
