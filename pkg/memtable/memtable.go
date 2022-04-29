package memtable

import (
	"encoding/binary"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/coocood/freecache"
)

// MemTable ...
type MemTable struct {
	cache *freecache.Cache
}

var _ dhash.MemTable = &MemTable{}

// New creates freecache with size
func New(size int) *MemTable {
	return &MemTable{
		cache: freecache.NewCache(size),
	}
}

// GetNum ...
func (m *MemTable) GetNum(key string) (num uint64, ok bool) {
	data, err := m.cache.Get([]byte(key))
	if err != nil {
		return 0, false
	}
	if len(data) < 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint64(data), true
}

// SetNum ...
func (m *MemTable) SetNum(key string, num uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], num)
	_ = m.cache.Set([]byte(key), data[:], 0)
}
