package repository

import (
	"context"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"sort"
)

// HashDatabase ...
type HashDatabase struct {
	doFetchSizeLog  func(ctx context.Context) (uint64, error)
	doSelectEntries func(ctx context.Context, inputs []HashRange) ([]dhash.Entry, error)

	fetchNew bool

	needFetchSizeLog bool
	sizeLog          uint64

	selectInputs  []HashRange
	outputEntries []dhash.Entry

	err error
}

// NewHashDatabase ...
func NewHashDatabase(
	fetchSizeLog func(ctx context.Context) (uint64, error),
	selectEntries func(ctx context.Context, inputs []HashRange) ([]dhash.Entry, error),
) *HashDatabase {
	return &HashDatabase{
		doFetchSizeLog:  fetchSizeLog,
		doSelectEntries: selectEntries,
	}
}

func (h *HashDatabase) fetchData(ctx context.Context) error {
	if h.err != nil {
		return h.err
	}
	h.err = h.fetchDataWithError(ctx)
	return h.err
}

func (h *HashDatabase) fetchDataWithError(ctx context.Context) error {
	if !h.fetchNew {
		return h.err
	}
	h.fetchNew = false

	if h.needFetchSizeLog {
		h.needFetchSizeLog = false

		sizeLog, err := h.doFetchSizeLog(ctx)
		if err != nil {
			return err
		}
		h.sizeLog = sizeLog
	}

	if len(h.selectInputs) > 0 {
		inputs := h.selectInputs
		h.selectInputs = nil

		entries, err := h.doSelectEntries(ctx, inputs)
		if err != nil {
			return err
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Hash < entries[j].Hash
		})
		h.outputEntries = entries
	}

	return nil
}

// GetSizeLog ...
func (h *HashDatabase) GetSizeLog(ctx context.Context) func() (uint64, error) {
	h.fetchNew = true
	h.needFetchSizeLog = true
	return func() (uint64, error) {
		if err := h.fetchData(ctx); err != nil {
			return 0, err
		}
		return h.sizeLog, nil
	}
}

// SelectEntries ...
func (h *HashDatabase) SelectEntries(
	ctx context.Context, hashBegin uint32, hashEnd dhash.NullUint32,
) func() ([]dhash.Entry, error) {
	h.fetchNew = true

	h.selectInputs = append(h.selectInputs, HashRange{
		Begin: hashBegin,
		End:   hashEnd,
	})

	return func() ([]dhash.Entry, error) {
		if err := h.fetchData(ctx); err != nil {
			return nil, err
		}

		var result []dhash.Entry
		for _, e := range h.outputEntries {
			if e.Hash < hashBegin {
				continue
			}
			if hashEnd.Valid && e.Hash >= hashEnd.Num {
				continue
			}
			result = append(result, e)
		}
		return result, nil
	}
}
