package model

import "time"

// Event ...
type Event struct {
	ID   uint64 `db:"id"`
	Seq  uint32 `db:"seq"`
	Data []byte `db:"data"`

	AggregateType AggregateType `db:"aggregate_type"`
	AggregateID   uint32        `db:"aggregate_id"`

	CreatedAt time.Time `db:"created_at"`
}

// AggregateType ...
type AggregateType int

const (
	// AggregateTypeBlacklist ...
	AggregateTypeBlacklist AggregateType = 1

	// AggregateTypeCampaign ...
	AggregateTypeCampaign AggregateType = 2
)
