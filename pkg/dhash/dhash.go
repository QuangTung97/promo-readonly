package dhash

import (
	"context"
	"time"
)

//go:generate moq -out dash_mocks_test.go . MemTable CacheClient CachePipeline HashDatabase

// MemTable for in memory hash table storing size log (with eviction)
type MemTable interface {
	// GetNum may not return entry that it just set
	GetNum(key string) (num uint64, ok bool)
	SetNum(key string, num uint64)
}

// LeaseGetType ...
type LeaseGetType int

const (
	// LeaseGetTypeOK when entry is found
	LeaseGetTypeOK LeaseGetType = 1

	// LeaseGetTypeGranted when entry is not found but lease is granted
	LeaseGetTypeGranted LeaseGetType = 2

	// LeaseGetTypeRejected when entry is not found and lease is not granted
	LeaseGetTypeRejected LeaseGetType = 3
)

// GetOutput ...
type GetOutput struct {
	Found bool
	Data  []byte
}

// LeaseGetOutput ...
type LeaseGetOutput struct {
	Type    LeaseGetType
	Data    []byte
	LeaseID uint64
}

// CacheClient for remote cache (like memcached)
type CacheClient interface {
	// Pipeline can NOT be shared between goroutines
	Pipeline() CachePipeline
}

// Timer ...
type Timer interface {
	Now() time.Time
	Sleep(d time.Duration)
}

// CachePipeline for batching cache requests
type CachePipeline interface {
	Get(key string) func() (GetOutput, error)
	LeaseGet(key string) func() (LeaseGetOutput, error)
	LeaseSet(key string, value []byte, leaseID uint64, ttl uint32) func() error
	Delete(key string) func() error
	Finish()
}

// Entry for single hash value entry
type Entry struct {
	Hash uint32
	Data []byte // marshalled record
}

// NullUint32 ...
type NullUint32 struct {
	Valid bool
	Num   uint32
}

// HashDatabase for the backing store
type HashDatabase interface {
	GetSizeLog(ctx context.Context) func() (uint64, error)
	SelectEntries(ctx context.Context, hashBegin uint32, hashEnd NullUint32) func() ([]Entry, error)
}

// Provider can be shared between goroutines
type Provider interface {
	NewSession() Session
}

// Session can NOT be shared between goroutines
type Session interface {
	NewHash(namespace string, db HashDatabase) Hash
}

// Hash ...
type Hash interface {
	SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error)
	InvalidateSizeLog(ctx context.Context) func() error
	InvalidateEntry(ctx context.Context, hash uint32) func() error
}

// NewProvider ...
func NewProvider(mem MemTable, client CacheClient, timer Timer) Provider {
	return &providerImpl{
		mem:    mem,
		client: client,
		timer:  timer,
	}
}

type providerImpl struct {
	mem    MemTable
	client CacheClient
	timer  Timer
}

type delayedCall struct {
	startedAt time.Time
	call      func()
}

type sessionImpl struct {
	mem      MemTable
	pipeline CachePipeline
	timer    Timer

	nextCalls []func()
	delayed   delayedCallHeap
}

func (s *sessionImpl) addNextCall(fn func()) {
	s.nextCalls = append(s.nextCalls, fn)
}

func (s *sessionImpl) addDelayedCall(startedAt time.Time, call func()) {
	s.delayed.push(delayedCall{
		startedAt: startedAt,
		call:      call,
	})
}

func (s *sessionImpl) processAllCalls() {
	for {
		for len(s.nextCalls) > 0 {
			nextCalls := s.nextCalls
			s.nextCalls = nil

			for _, call := range nextCalls {
				call()
			}
		}

		if s.delayed.size() == 0 {
			return
		}

		now := s.timer.Now()

		top := s.delayed.pop()
		sleepDuration := top.startedAt.Sub(now)
		s.timer.Sleep(sleepDuration)

		top.call()

		// now >= startedAt <=> ~(now < startedAt)
		for s.delayed.size() > 0 && !now.Before(s.delayed.top().startedAt) {
			top := s.delayed.pop()
			top.call()
		}
	}
}

type hashImpl struct {
	sess *sessionImpl

	mem        MemTable
	pipeline   CachePipeline
	db         HashDatabase
	namespace  string
	sizeLogKey string
}

// NewSession ...
func (p *providerImpl) NewSession() Session {
	return &sessionImpl{
		mem:      p.mem,
		pipeline: p.client.Pipeline(),
		timer:    p.timer,
	}
}

// New ...
func (s *sessionImpl) NewHash(namespace string, db HashDatabase) Hash {
	return &hashImpl{
		sess: s,

		mem:        s.mem,
		pipeline:   s.pipeline,
		db:         db,
		namespace:  namespace,
		sizeLogKey: namespace + ":size-log",
	}
}

// SelectEntries ...
func (h *hashImpl) SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error) {
	action := &hashSelectAction{
		root: h,
		ctx:  ctx,
		hash: hash,
	}

	sizeLogNum, ok := h.mem.GetNum(h.namespace)
	if !ok {
		action.getSizeLogFromClient()

		h.sess.addNextCall(func() {
			action.handleMemSizeLogNotExisted()
		})
	} else {
		sizeLog := int(sizeLogNum)
		action.sizeLog = sizeLog

		action.getSizeLogFromClient()
		action.getBuckets()
		h.sess.addNextCall(func() {
			action.handleMemSizeLogExisted()
		})
	}

	return func() ([]Entry, error) {
		h.sess.processAllCalls()
		return action.results, action.err
	}
}

// InvalidateSizeLog ...
func (h *hashImpl) InvalidateSizeLog(_ context.Context) func() error {
	return func() error {
		return nil
	}
}

// InvalidateEntry ...
func (h *hashImpl) InvalidateEntry(_ context.Context, _ uint32) func() error {
	return func() error {
		return nil
	}
}
