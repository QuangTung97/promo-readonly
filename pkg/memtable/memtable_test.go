package memtable

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemTable(t *testing.T) {
	m := New(16 * 1024)

	m.SetNum("key01", 11)
	m.SetNum("key02", 12)

	n, ok := m.GetNum("key01")
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(11), n)

	n, ok = m.GetNum("key02")
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(12), n)

	n, ok = m.GetNum("key03")
	assert.Equal(t, false, ok)
	assert.Equal(t, uint64(0), n)

	_ = m.cache.Set([]byte("key04"), []byte("aa"), 0)
	n, ok = m.GetNum("key04")
	assert.Equal(t, false, ok)
	assert.Equal(t, uint64(0), n)
}
