package dhash

import (
	"context"
	"fmt"
)

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

// Database for the backing store
type Database interface {
	GetSizeLog(ctx context.Context) func() (uint64, error)
	SelectEntries(ctx context.Context, hashBegin uint32, hashEnd uint32) func() ([]Entry, error)
}

// Provider can be shared between goroutines
type Provider interface {
	NewSession() Session
}

// Session can NOT be shared between goroutines
type Session interface {
	New(namespace string, db Database) Hash
}

// Hash ...
type Hash interface {
	SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error)
	InvalidateSizeLog(ctx context.Context) func() error
	InvalidateEntry(ctx context.Context, hash uint32) func() error
}

// NewProvider ...
func NewProvider(mem MemTable, client CacheClient) Provider {
	return &providerImpl{
		mem:    mem,
		client: client,
	}
}

type providerImpl struct {
	mem    MemTable
	client CacheClient
}

type sessionImpl struct {
	mem      MemTable
	pipeline CachePipeline
}

type hashImpl struct {
	sess *sessionImpl

	mem        MemTable
	pipeline   CachePipeline
	db         Database
	namespace  string
	sizeLogKey string
}

// NewSession ...
func (p *providerImpl) NewSession() Session {
	return &sessionImpl{
		mem:      p.mem,
		pipeline: p.client.Pipeline(),
	}
}

// New ...
func (s *sessionImpl) New(namespace string, db Database) Hash {
	return &hashImpl{
		mem:        s.mem,
		pipeline:   s.pipeline,
		db:         db,
		namespace:  namespace,
		sizeLogKey: namespace + ":size-log",
	}
}

// SelectEntries ...
func (h *hashImpl) SelectEntries(ctx context.Context, hash uint32) func() ([]Entry, error) {
	sizeLog, ok := h.mem.GetNum(h.namespace)
	if !ok {

	}

	sizeLogFn := h.pipeline.LeaseGet(h.sizeLogKey)
	bucketFn1 := h.pipeline.Get(fmt.Sprintf("%s:%d:%d", h.namespace, sizeLog-1, hash))
	bucketFn2 := h.pipeline.Get(fmt.Sprintf("%s:%d:%d", h.namespace, sizeLog, hash))

	return func() ([]Entry, error) {
		newSizeLogOutput, err := sizeLogFn()
		if err != nil {
			return nil, err
		}
		// TODO Handle Size Log Different
		if newSizeLogOutput.Type == LeaseGetTypeOK {
		}

		bucket1Output, err := bucketFn1()
		if err != nil {
			return nil, err
		}
		// TODO
		if !bucket1Output.Found {
		}

		bucket2Output, err := bucketFn2()
		if err != nil {
			return nil, err
		}

		if !bucket1Output.Found && !bucket2Output.Found {
		}

		return unmarshalEntries(bucket2Output.Data)
	}
}

// InvalidateSizeLog ...
func (h *hashImpl) InvalidateSizeLog(ctx context.Context) func() error {
	return func() error {
		return nil
	}
}

// InvalidateEntry ...
func (h *hashImpl) InvalidateEntry(ctx context.Context, hash uint32) func() error {
	return func() error {
		return nil
	}
}
