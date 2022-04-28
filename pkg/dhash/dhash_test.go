package dhash

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type hashTest struct {
	mem   *MemTableMock
	db    *HashDatabaseMock
	pipe  *CachePipelineMock
	hash  Hash
	timer *timerMock
}

type timerMock struct {
	nowCalls   int
	current    time.Time
	sleepCalls []time.Duration
}

func (t *timerMock) Now() time.Time {
	return t.current
}

func (t *timerMock) Sleep(d time.Duration) {
	t.sleepCalls = append(t.sleepCalls, d)
	t.current = t.current.Add(d)
}

func startOfTime() time.Time {
	return newTime("2022-05-07T10:00:00+07:00")
}

func newTimeMock() *timerMock {
	m := &timerMock{
		nowCalls: 0,
		current:  startOfTime(),
	}
	return m
}

func newHashTest(ns string, options ...SessionOption) *hashTest {
	mem := &MemTableMock{}
	client := &CacheClientMock{}
	pipeline := &CachePipelineMock{}

	client.PipelineFunc = func() CachePipeline {
		return pipeline
	}

	timer := newTimeMock()
	p := newProviderImpl(mem, client)
	p.timer = timer

	db := &HashDatabaseMock{}

	h := &hashTest{
		mem:   mem,
		db:    db,
		pipe:  pipeline,
		hash:  p.NewSession(options...).NewHash(ns, db),
		timer: timer,
	}

	h.stubMemTable()
	h.stubPipeline()
	h.stubDB()
	return h
}

func (h *hashTest) stubMemTable() {
	h.mem.GetNumFunc = func(key string) (uint64, bool) {
		return 0, false
	}
	h.mem.SetNumFunc = func(key string, num uint64) {}
}

func (h *hashTest) stubPipeline() {
	h.pipe.GetFunc = func(key string) func() (GetOutput, error) {
		return func() (GetOutput, error) {
			return GetOutput{
				Found: true,
				Data:  []byte("default-data"),
			}, nil
		}
	}
	h.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		return func() (LeaseGetOutput, error) {
			return LeaseGetOutput{
				Type: LeaseGetTypeOK,
				Data: []byte("default-data"),
			}, nil
		}
	}
	h.pipe.LeaseSetFunc = func(key string, value []byte, leaseID uint64, ttl uint32) func() error {
		return func() error {
			return nil
		}
	}
}

func (h *hashTest) stubDB() {
	h.db.GetSizeLogFunc = func(ctx context.Context) func() (uint64, error) {
		return func() (uint64, error) {
			return 0, nil
		}
	}
	h.db.SelectEntriesFunc = func(ctx context.Context, hashBegin uint32, hashEnd NullUint32) func() ([]Entry, error) {
		return func() ([]Entry, error) {
			return nil, nil
		}
	}
}

func (h *hashTest) stubGetNum(num uint64) {
	h.mem.GetNumFunc = func(key string) (uint64, bool) {
		return num, true
	}
}

func (h *hashTest) stubLeaseGet(output LeaseGetOutput) {
	h.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		return func() (LeaseGetOutput, error) {
			return output, nil
		}
	}
}

func (h *hashTest) stubLeaseGetOK(data string) {
	h.stubLeaseGet(LeaseGetOutput{
		Type: LeaseGetTypeOK,
		Data: []byte(data),
	})
}

func (h *hashTest) stubLeaseGetOutputs(outputs []LeaseGetOutput) {
	h.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		index := len(h.pipe.LeaseGetCalls()) - 1
		return func() (LeaseGetOutput, error) {
			return outputs[index], nil
		}
	}
}

func (h *hashTest) stubGetNumNotFound() {
	h.mem.GetNumFunc = func(key string) (uint64, bool) {
		return 0, false
	}
}

func (h *hashTest) stubDBGetSizeLog(n uint64) {
	h.db.GetSizeLogFunc = func(ctx context.Context) func() (uint64, error) {
		return func() (uint64, error) {
			return n, nil
		}
	}
}

func (h *hashTest) stubDBSelectEntries(entries []Entry) {
	h.db.SelectEntriesFunc = func(ctx context.Context, hashBegin uint32, hashEnd NullUint32) func() ([]Entry, error) {
		return func() ([]Entry, error) {
			return entries, nil
		}
	}
}

func (h *hashTest) stubClientGet(entriesList [][]Entry) {
	h.pipe.GetFunc = func(key string) func() (GetOutput, error) {
		index := len(h.pipe.GetCalls()) - 1

		return func() (GetOutput, error) {
			entries := entriesList[index]
			if len(entries) == 0 {
				return GetOutput{}, nil
			}
			return GetOutput{
				Found: true,
				Data:  marshalEntries(entries),
			}, nil
		}
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

func newEntry(hash uint32, data ...byte) Entry {
	return Entry{Hash: hash, Data: data}
}

func TestSelectEntries__Second_Slot_Found(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOK("5")
	h.stubClientGet([][]Entry{
		{},
		{
			newEntry(0xfc345678, 1, 2, 3),
			newEntry(0xfc345000, 5, 6, 7),
		},
	})

	entries, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{newEntry(0xfc345678, 1, 2, 3)}, entries)
}

func TestSelectEntries__First_Slot_Found(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOK("5")
	h.stubClientGet([][]Entry{
		{
			newEntry(0xfc345678, 1, 2, 3),
			newEntry(0xfc345000, 5, 6, 7),
		},
		{},
	})

	entries, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{newEntry(0xfc345678, 1, 2, 3)}, entries)
}

func TestSelectEntries__When_Call_Get_Num_Not_Found__Only_Call_LeaseGet_Size_Log(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNumNotFound()

	h.hash.SelectEntries(newContext(), 123)

	assert.Equal(t, 1, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, 0, len(h.pipe.GetCalls()))
}

func TestSelectEntries__When_GetNum_Not_Found__Call_Client_Get(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNumNotFound()
	h.stubLeaseGet(LeaseGetOutput{
		Type: LeaseGetTypeOK,
		Data: []byte("5"),
	})

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()

	assert.Equal(t, 2, len(h.pipe.GetCalls()))
	assert.Equal(t, "sample:4:f0000000", h.pipe.GetCalls()[0].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.GetCalls()[1].Key)
}

func TestSelectEntries__When_GetNum_Not_Found__Second_Slot_Found(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNumNotFound()
	h.stubLeaseGetOK("5")
	h.stubClientGet([][]Entry{
		{},
		{
			newEntry(0xfc345678, 1, 2, 3),
			newEntry(0xfc345000, 5, 6, 7),
		},
	})

	entries, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{newEntry(0xfc345678, 1, 2, 3)}, entries)
}

func TestSelectEntries__When_Client_Get_Size_Log_Granted__Call_Get_Size_Log_From_DB(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGet(LeaseGetOutput{
		Type:    LeaseGetTypeGranted,
		LeaseID: 0x3344,
	})

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, 1, len(h.db.GetSizeLogCalls()))
	assert.Equal(t, newContext(), h.db.GetSizeLogCalls()[0].Ctx)
}

func TestSelectEntries__When_Client_Get_Size_Log_Granted__Do_MemTable_SetNum(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGet(LeaseGetOutput{
		Type:    LeaseGetTypeGranted,
		LeaseID: 0x3344,
	})

	h.stubDBGetSizeLog(7)

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()

	assert.Equal(t, 1, len(h.mem.SetNumCalls()))
	assert.Equal(t, "sample", h.mem.SetNumCalls()[0].Key)
	assert.Equal(t, uint64(7), h.mem.SetNumCalls()[0].Num)
}

func TestSelectEntries__When_Client_Get_Size_Log_Granted__Do_Cache_Client_Lease_Set(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGet(LeaseGetOutput{
		Type:    LeaseGetTypeGranted,
		LeaseID: 0x3344,
	})

	h.stubDBGetSizeLog(7)

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()

	assert.Equal(t, 1, len(h.pipe.LeaseSetCalls()))
	assert.Equal(t, "sample:size-log", h.pipe.LeaseSetCalls()[0].Key)
	assert.Equal(t, []byte("7"), h.pipe.LeaseSetCalls()[0].Value)
	assert.Equal(t, uint64(0x3344), h.pipe.LeaseSetCalls()[0].LeaseID)
	assert.Equal(t, uint32(0), h.pipe.LeaseSetCalls()[0].TTL)
}

func TestSelectEntries__When_Client_Get_Size_Log_Granted__Returns_Entry_From_Client_Get(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGet(LeaseGetOutput{
		Type:    LeaseGetTypeGranted,
		LeaseID: 0x3344,
	})
	h.stubClientGet([][]Entry{
		{},
		{
			newEntry(0xfc345678, 1, 2, 3),
		},
	})

	h.stubDBGetSizeLog(7)

	entries, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{
		newEntry(0xfc345678, 1, 2, 3),
	}, entries)
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Client_Lease_Get(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOK("5")
	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()

	assert.Equal(t, 2, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, "sample:size-log", h.pipe.LeaseGetCalls()[0].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[1].Key)
}

func newLeaseGetGranted(leaseID uint64) LeaseGetOutput {
	return LeaseGetOutput{
		Type:    LeaseGetTypeGranted,
		LeaseID: leaseID,
	}
}

func newLeaseGetRejected() LeaseGetOutput {
	return LeaseGetOutput{
		Type: LeaseGetTypeRejected,
	}
}

func newNullUint32(v uint32) NullUint32 {
	return NullUint32{
		Valid: true,
		Num:   v,
	}
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Select_Entries_From_DB(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOutputs([]LeaseGetOutput{
		{
			Type: LeaseGetTypeOK,
			Data: []byte("5"),
		},
		newLeaseGetGranted(7788),
	})

	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	_, _ = h.hash.SelectEntries(newContext(), 0xdc345678)()

	assert.Equal(t, 1, len(h.db.SelectEntriesCalls()))
	assert.Equal(t, uint32(0xd8000000), h.db.SelectEntriesCalls()[0].HashBegin)
	assert.Equal(t, newNullUint32(0xe0000000), h.db.SelectEntriesCalls()[0].HashEnd)
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Returns_Entry_From_DB__And_Set_ClientCache(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOutputs([]LeaseGetOutput{
		{
			Type: LeaseGetTypeOK,
			Data: []byte("5"),
		},
		newLeaseGetGranted(7788),
	})

	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	dbEntries := []Entry{
		{
			Hash: 0xdc345679,
			Data: []byte("db data 01"),
		},
		{
			Hash: 0xdc345679,
			Data: []byte("db data 02"),
		},
	}
	h.stubDBSelectEntries(dbEntries)

	entries, err := h.hash.SelectEntries(newContext(), 0xdc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{
		{
			Hash: 0xdc345679,
			Data: []byte("db data 01"),
		},
		{
			Hash: 0xdc345679,
			Data: []byte("db data 02"),
		},
	}, entries)

	assert.Equal(t, 1, len(h.pipe.LeaseSetCalls()))
	assert.Equal(t, "sample:5:d8000000", h.pipe.LeaseSetCalls()[0].Key)
	assert.Equal(t, marshalEntries(dbEntries), h.pipe.LeaseSetCalls()[0].Value)
	assert.Equal(t, uint64(7788), h.pipe.LeaseSetCalls()[0].LeaseID)
	assert.Equal(t, uint32(0), h.pipe.LeaseSetCalls()[0].TTL)
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Client_Lease_Get_Rejected__Call_Second_Times(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOutputs([]LeaseGetOutput{
		{
			Type: LeaseGetTypeOK,
			Data: []byte("5"),
		},
		newLeaseGetRejected(),
		newLeaseGetGranted(5544),
	})
	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	_, _ = h.hash.SelectEntries(newContext(), 0xfc345678)()

	assert.Equal(t, 3, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[1].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[2].Key)

	assert.Equal(t, []time.Duration{
		10 * time.Millisecond,
	}, h.timer.sleepCalls)
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Client_Lease_Get_Rejected_All_Times__Returns_Err(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOutputs([]LeaseGetOutput{
		{
			Type: LeaseGetTypeOK,
			Data: []byte("5"),
		},
		newLeaseGetRejected(),
		newLeaseGetRejected(),
		newLeaseGetRejected(),
		newLeaseGetRejected(),
	})
	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	_, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, ErrLeaseNotGranted, err)

	assert.Equal(t, 5, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[1].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[2].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[3].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[4].Key)

	// default durations
	assert.Equal(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
	}, h.timer.sleepCalls)
}

func TestSelectEntries__When_Both_Bucket_Not_Found__Client_Lease_Get_OK__Returns_Client_Entries(t *testing.T) {
	h := newHashTest("sample")

	h.stubGetNum(5)
	h.stubLeaseGetOutputs([]LeaseGetOutput{
		{
			Type: LeaseGetTypeOK,
			Data: []byte("5"),
		},
		newLeaseGetRejected(),
		{
			Type: LeaseGetTypeOK,
			Data: marshalEntries([]Entry{
				{
					Hash: 0xfc345678,
					Data: []byte("sample data"),
				},
			}),
		},
	})
	h.stubClientGet([][]Entry{
		{}, {}, // both not found
	})

	entries, err := h.hash.SelectEntries(newContext(), 0xfc345678)()
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{
		{
			Hash: 0xfc345678,
			Data: []byte("sample data"),
		},
	}, entries)

	assert.Equal(t, 3, len(h.pipe.LeaseGetCalls()))
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[1].Key)
	assert.Equal(t, "sample:5:f8000000", h.pipe.LeaseGetCalls()[2].Key)

	assert.Equal(t, []time.Duration{
		10 * time.Millisecond,
	}, h.timer.sleepCalls)

	assert.Equal(t, 0, len(h.db.SelectEntriesCalls()))
}
