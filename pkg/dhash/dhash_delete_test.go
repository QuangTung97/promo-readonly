package dhash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash_InvalidateSizeLog__Call_Delete_On_Client(t *testing.T) {
	h := newHashTest("sample")

	err := h.hash.InvalidateSizeLog(newContext())()
	assert.Equal(t, nil, err)

	assert.Equal(t, 1, len(h.pipe.DeleteCalls()))
	assert.Equal(t, "sample:size-log", h.pipe.DeleteCalls()[0].Key)
}

func TestHash_InvalidateEntry__Call_Delete_On_2_Buckets(t *testing.T) {
	h := newHashTest("sample")

	err := h.hash.InvalidateEntry(newContext(), 4, 0xfc345678)()
	assert.Equal(t, nil, err)

	assert.Equal(t, 2, len(h.pipe.DeleteCalls()))
	assert.Equal(t, "sample:3:e0000000", h.pipe.DeleteCalls()[0].Key)
	assert.Equal(t, "sample:4:f0000000", h.pipe.DeleteCalls()[1].Key)
}
