package dhash

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

//go:generate moq -out dhash_mocks_test.go . MemTable CacheClient CachePipeline HashDatabase StoreDatabase

//go:generate moq -out dhash_mocks.go . Session Hash

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

type delayTimer interface {
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

// StoreDatabase backing store of simple kv cache
type StoreDatabase interface {
	Get(ctx context.Context, key string) func() ([]byte, error)
}

// Provider can be shared between goroutines
type Provider interface {
	NewSession(options ...SessionOption) Session
}

// Session can NOT be shared between goroutines
type Session interface {
	NewHash(namespace string, db HashDatabase) Hash
	NewStore(fn StoreDatabase) Store
	Finish()
}

// ErrLeaseNotGranted after multiple retries configured by WithWaitLeaseDurations
var ErrLeaseNotGranted = errors.New("lease not granted after retries")

// Hash likes Redis hash map (but consistent)
type Hash interface {
	SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error)
	InvalidateSizeLog(ctx context.Context) func() error
	InvalidateEntry(ctx context.Context, sizeLog uint64, hash uint32) func() error
}

// Store for simple kv store (plain memcached key-value)
type Store interface {
	Get(ctx context.Context, key string) func() ([]byte, error)
	Invalidate(ctx context.Context, key string) func() error
}

type defaultDelayTimer struct {
}

func (t defaultDelayTimer) Now() time.Time {
	return time.Now()
}

func (t defaultDelayTimer) Sleep(d time.Duration) {
	time.Sleep(d)
}

func newProviderImpl(mem MemTable, client CacheClient) *ProviderImpl {
	return &ProviderImpl{
		mem:    mem,
		client: client,
		timer:  defaultDelayTimer{},
	}
}

// NewProvider ...
func NewProvider(mem MemTable, client CacheClient) *ProviderImpl {
	return newProviderImpl(mem, client)
}

// ProviderImpl ...
type ProviderImpl struct {
	mem    MemTable
	client CacheClient
	timer  delayTimer

	hashSizeLogAccessCount uint64
	hashBucketAccessCount  uint64

	hashSizeLogMissCount uint64
	hashBucketMissCount  uint64

	storeAccessCount uint64
	storeMissCount   uint64
}

// HashSizeLogAccessCount ...
func (p *ProviderImpl) HashSizeLogAccessCount() uint64 {
	return atomic.LoadUint64(&p.hashSizeLogAccessCount)
}

// HashBucketAccessCount ...
func (p *ProviderImpl) HashBucketAccessCount() uint64 {
	return atomic.LoadUint64(&p.hashBucketAccessCount)
}

// HashSizeLogMissCount ...
func (p *ProviderImpl) HashSizeLogMissCount() uint64 {
	return atomic.LoadUint64(&p.hashSizeLogMissCount)
}

// HashBucketMissCount ...
func (p *ProviderImpl) HashBucketMissCount() uint64 {
	return atomic.LoadUint64(&p.hashBucketMissCount)
}

type delayedCall struct {
	startedAt time.Time
	call      func()
}

type sessionImpl struct {
	provider *ProviderImpl

	options sessionOptions

	mem      MemTable
	pipeline *deduplicatedPipeline
	timer    delayTimer

	nextCalls []func()
	delayed   delayedCallHeap

	hashSizeLogAccessCount uint64
	hashBucketAccessCount  uint64

	hashSizeLogMissCount uint64
	hashBucketMissCount  uint64

	storeAccessCount uint64
	storeMissCount   uint64
}

func (s *sessionImpl) addNextCall(fn func()) {
	s.nextCalls = append(s.nextCalls, fn)
}

func (s *sessionImpl) addDelayedCall(d time.Duration, call func()) {
	s.delayed.push(delayedCall{
		startedAt: s.timer.Now().Add(d),
		call:      call,
	})
}

func (s *sessionImpl) processAllCalls() {
	for {
		for len(s.nextCalls) > 0 {
			s.pipeline.reset()

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

		s.pipeline.reset()
		top.call()

		now = s.timer.Now().Add(200 * time.Microsecond) // earlier about 200 microseconds

		// now >= startedAt <=> ~(now < startedAt)
		for s.delayed.size() > 0 && !now.Before(s.delayed.top().startedAt) {
			top := s.delayed.pop()
			top.call()
		}
	}
}

// NewSession ...
func (p *ProviderImpl) NewSession(options ...SessionOption) Session {
	return &sessionImpl{
		provider: p,
		options:  newSessionOptions(options...),
		mem:      p.mem,
		pipeline: newDeduplicatedPipeline(p.client.Pipeline()),
		timer:    p.timer,
	}
}

// NewHash ...
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

// NewStore ...
func (s *sessionImpl) NewStore(db StoreDatabase) Store {
	return &storeImpl{
		sess:     s,
		db:       db,
		pipeline: s.pipeline,
	}
}

// Finish ...
func (s *sessionImpl) Finish() {
	atomic.AddUint64(&s.provider.hashSizeLogAccessCount, s.hashSizeLogAccessCount)
	s.hashSizeLogAccessCount = 0

	atomic.AddUint64(&s.provider.hashBucketAccessCount, s.hashBucketAccessCount)
	s.hashBucketAccessCount = 0

	atomic.AddUint64(&s.provider.hashSizeLogMissCount, s.hashSizeLogMissCount)
	s.hashSizeLogMissCount = 0

	atomic.AddUint64(&s.provider.hashBucketMissCount, s.hashBucketMissCount)
	s.hashBucketMissCount = 0

	s.pipeline.Finish()
}
