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

func newBlacklistCustomerHashDB(repo repository.Blacklist) dhash.HashDatabase {
	return repository.NewHashDatabase(func(ctx context.Context) (uint64, error) {
		config, err := repo.GetConfig(ctx)
		if err != nil {
			return 0, err
		}
		return log2Int(config.CustomerCount), nil
	}, func(ctx context.Context, inputs []repository.HashRange) ([]dhash.Entry, error) {
		customers, err := repo.SelectBlacklistCustomers(ctx, inputs)
		if err != nil {
			return nil, err
		}

		entries := make([]dhash.Entry, 0, len(customers))
		for _, c := range customers {
			entries = append(entries, dhash.Entry{
				Hash: c.Hash,
				Data: marshalBlacklistCustomer(c),
			})
		}
		return entries, nil
	})
}

func newCampaignHashDB(_ repository.Campaign) dhash.HashDatabase {
	return nil
}

func marshalBlacklistMerchant(m model.BlacklistMerchant) []byte {
	msg := promopb.BlacklistMerchantData{
		Hash:         m.Hash,
		MerchantCode: m.MerchantCode,
		Status:       uint32(m.Status),
		StartTime:    newTimestampNull(m.StartTime),
		EndTime:      newTimestampNull(m.EndTime),
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	return data
}

func unmarshalBlacklistMerchant(data []byte) (model.BlacklistMerchant, error) {
	var msg promopb.BlacklistMerchantData
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return model.BlacklistMerchant{}, err
	}
	return model.BlacklistMerchant{
		Hash:         msg.Hash,
		MerchantCode: msg.MerchantCode,
		Status:       model.BlacklistMerchantStatus(msg.Status),
		StartTime:    nullTimeFromTimestamp(msg.StartTime),
		EndTime:      nullTimeFromTimestamp(msg.EndTime),
	}, nil
}

func newBlacklistMerchantHashDB(repo repository.Blacklist) dhash.HashDatabase {
	return repository.NewHashDatabase(func(ctx context.Context) (uint64, error) {
		config, err := repo.GetConfig(ctx)
		if err != nil {
			return 0, err
		}
		return log2Int(config.MerchantCount), nil
	}, func(ctx context.Context, inputs []repository.HashRange) ([]dhash.Entry, error) {
		merchants, err := repo.SelectBlacklistMerchants(ctx, inputs)
		if err != nil {
			return nil, err
		}

		entries := make([]dhash.Entry, 0, len(merchants))
		for _, m := range merchants {
			entries = append(entries, dhash.Entry{
				Hash: m.Hash,
				Data: marshalBlacklistMerchant(m),
			})
		}
		return entries, nil
	})
}
