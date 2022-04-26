package dhash

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type hashTest struct {
	mem  *MemTableMock
	db   *DatabaseMock
	pipe *CachePipelineMock
	hash Hash
}

func newHashTest(ns string) *hashTest {
	mem := &MemTableMock{}
	client := &CacheClientMock{}
	pipeline := &CachePipelineMock{}

	client.PipelineFunc = func() CachePipeline {
		return pipeline
	}

	p := NewProvider(mem, client)

	db := &DatabaseMock{}

	h := &hashTest{
		mem:  mem,
		db:   db,
		pipe: pipeline,
		hash: p.NewSession().New(ns, db),
	}

	h.stubMemTable()
	h.stubPipeline()
	return h
}

func (h *hashTest) stubMemTable() {
	h.mem.GetNumFunc = func(key string) (uint64, bool) {
		return 0, false
	}
}

func (h *hashTest) stubPipeline() {
	h.pipe.GetFunc = func(key string) func() (GetOutput, error) {
		return func() (GetOutput, error) {
			return GetOutput{
				Found: false,
			}, nil
		}
	}
	h.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		return func() (LeaseGetOutput, error) {
			return LeaseGetOutput{
				Type: LeaseGetTypeRejected,
			}, nil
		}
	}
}

func (h *hashTest) stubGetNum(num uint64) {
	h.mem.GetNumFunc = func(key string) (uint64, bool) {
		return num, true
	}
}

func newContext() context.Context {
	return context.Background()
}

func TestSelectEntries_Call_Get_MemTable_Entry(t *testing.T) {
	h := newHashTest("sample")
	h.hash.SelectEntries(newContext(), 123)

	assert.Equal(t, 1, len(h.mem.GetNumCalls()))
	assert.Equal(t, "sample", h.mem.GetNumCalls()[0].Key)
}

func TestSelectEntries__MemTable_Exist__Call_Get_From_Cache_Client(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.hash.SelectEntries(newContext(), 0xfc345678)

	assert.Equal(t, 1, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, "sample:size-log", h.pipe.LeaseGetCalls()[0].Key)

	assert.Equal(t, 2, len(h.pipe.GetCalls()))
	assert.Equal(t, "sample:4:f0000000", h.pipe.GetCalls()[0].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.GetCalls()[1].Key)
}
