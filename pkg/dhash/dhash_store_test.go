package dhash

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type storeTest struct {
	pipe  *CachePipelineMock
	db    *StoreDatabaseMock
	store Store
	timer *timerMock
}

func newStoreTest() *storeTest {
	client := &CacheClientMock{}
	pipeline := &CachePipelineMock{}

	client.PipelineFunc = func() CachePipeline {
		return pipeline
	}

	timer := newTimeMock()
	p := newProviderImpl(nil, client)
	p.timer = timer

	db := &StoreDatabaseMock{}

	s := &storeTest{
		pipe:  pipeline,
		db:    db,
		store: p.NewSession().NewStore(db),
		timer: timer,
	}

	s.stubPipeline()
	s.stubDB()
	return s
}

func (s *storeTest) stubPipeline() {
	s.stubLeaseGet(LeaseGetOutput{
		Type: LeaseGetTypeOK,
		Data: []byte("default lease data return"),
	})
	s.pipe.LeaseSetFunc = func(key string, value []byte, leaseID uint64, ttl uint32) func() error {
		return func() error {
			return nil
		}
	}
	s.pipe.DeleteFunc = func(key string) func() error {
		return func() error {
			return nil
		}
	}
}

func (s *storeTest) stubDB() {
	s.stubDBGet("default db data")
}

func (s *storeTest) stubDBGet(data string) {
	s.db.GetFunc = func(ctx context.Context, key string) func() ([]byte, error) {
		return func() ([]byte, error) {
			return []byte(data), nil
		}
	}
}

func (s *storeTest) stubDBGetError(err error) {
	s.db.GetFunc = func(ctx context.Context, key string) func() ([]byte, error) {
		return func() ([]byte, error) {
			return nil, err
		}
	}
}

func (s *storeTest) stubDBGetList(dataList []string) {
	s.db.GetFunc = func(ctx context.Context, key string) func() ([]byte, error) {
		index := len(s.db.GetCalls()) - 1
		return func() ([]byte, error) {
			return []byte(dataList[index]), nil
		}
	}
}

func (s *storeTest) stubLeaseGet(output LeaseGetOutput) {
	s.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		return func() (LeaseGetOutput, error) {
			return output, nil
		}
	}
}

func (s *storeTest) stubLeaseGetOutputs(outputs []LeaseGetOutput) {
	s.pipe.LeaseGetFunc = func(key string) func() (LeaseGetOutput, error) {
		index := len(s.pipe.LeaseGetCalls()) - 1
		return func() (LeaseGetOutput, error) {
			return outputs[index], nil
		}
	}
}

func TestStore_Get__Call_Client_LeaseGet(t *testing.T) {
	s := newStoreTest()
	s.store.Get(newContext(), "key01")

	assert.Equal(t, 1, len(s.pipe.LeaseGetCalls()))
	assert.Equal(t, "key01", s.pipe.LeaseGetCalls()[0].Key)
}

func TestStore_Get__Lease_OK__Returns_Data(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGet(LeaseGetOutput{
		Type: LeaseGetTypeOK,
		Data: []byte("sample data"),
	})

	data, err := s.store.Get(newContext(), "key01")()

	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("sample data"), data)
}

func TestStore_Get__Lease_Granted__Call_Get_From_DB(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGet(newLeaseGetGranted(889900))

	_, _ = s.store.Get(newContext(), "key01")()

	assert.Equal(t, 1, len(s.db.GetCalls()))
	assert.Equal(t, newContext(), s.db.GetCalls()[0].Ctx)
	assert.Equal(t, "key01", s.db.GetCalls()[0].Key)
}

func TestStore_Get__Lease_Granted__Call_Lease_Set(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGet(newLeaseGetGranted(889900))
	s.stubDBGet("db get data")

	_, _ = s.store.Get(newContext(), "key01")()

	assert.Equal(t, 1, len(s.pipe.LeaseSetCalls()))
	assert.Equal(t, "key01", s.pipe.LeaseSetCalls()[0].Key)
	assert.Equal(t, []byte("db get data"), s.pipe.LeaseSetCalls()[0].Value)
	assert.Equal(t, uint64(889900), s.pipe.LeaseSetCalls()[0].LeaseID)
	assert.Equal(t, uint32(0), s.pipe.LeaseSetCalls()[0].TTL)
}

func TestStore_Get__Lease_Granted__Returns_Data(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGet(newLeaseGetGranted(889900))
	s.stubDBGet("db get data")

	data, err := s.store.Get(newContext(), "key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, "db get data", string(data))
}

func TestStore_Get__Lease_Granted__Get_From_DB_Err(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGet(newLeaseGetGranted(889900))
	s.stubDBGetError(errors.New("some error"))

	data, err := s.store.Get(newContext(), "key01")()
	assert.Equal(t, errors.New("some error"), err)
	assert.Nil(t, data)
}

func TestStore_Get__Lease_Rejected__Call_Lease_Get_Multiple_Times__Returns_Error(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGetOutputs([]LeaseGetOutput{
		newLeaseGetRejected(),
		newLeaseGetRejected(),
		newLeaseGetRejected(),
		newLeaseGetRejected(),
	})

	data, err := s.store.Get(newContext(), "key01")()
	assert.Equal(t, ErrLeaseNotGranted, err)
	assert.Nil(t, data)

	assert.Equal(t, 4, len(s.pipe.LeaseGetCalls()))
	assert.Equal(t, "key01", s.pipe.LeaseGetCalls()[0].Key)
	assert.Equal(t, "key01", s.pipe.LeaseGetCalls()[3].Key)

	assert.Equal(t, []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
	}, s.timer.sleepCalls)
}

func TestStore_Get__Lease_Rejected__Then_Granted(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGetOutputs([]LeaseGetOutput{
		newLeaseGetRejected(),
		newLeaseGetGranted(8833),
	})
	s.stubDBGet("some db data")

	data, err := s.store.Get(newContext(), "key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, "some db data", string(data))
}

func TestStore_Invalidate(t *testing.T) {
	s := newStoreTest()

	err := s.store.Invalidate(newContext(), "key01")()
	assert.Equal(t, nil, err)

	assert.Equal(t, 1, len(s.pipe.DeleteCalls()))
	assert.Equal(t, "key01", s.pipe.DeleteCalls()[0].Key)
}

func TestStore_Get__Rejected__Multi_Gets(t *testing.T) {
	s := newStoreTest()
	s.stubLeaseGetOutputs([]LeaseGetOutput{
		newLeaseGetRejected(),
		newLeaseGetRejected(),
		newLeaseGetGranted(3344),
		newLeaseGetGranted(5566),
	})
	s.stubDBGetList([]string{"db data 01", "db data 02"})

	fn1 := s.store.Get(newContext(), "key01")
	fn2 := s.store.Get(newContext(), "key02")

	data1, err := fn1()
	assert.Equal(t, nil, err)
	assert.Equal(t, "db data 01", string(data1))

	data2, err := fn2()
	assert.Equal(t, nil, err)
	assert.Equal(t, "db data 02", string(data2))

	assert.Equal(t, 2, len(s.db.GetCalls()))
	assert.Equal(t, "key01", s.db.GetCalls()[0].Key)
	assert.Equal(t, "key02", s.db.GetCalls()[1].Key)

	assert.Equal(t, []time.Duration{
		10 * time.Millisecond,
	}, s.timer.sleepCalls)
}
