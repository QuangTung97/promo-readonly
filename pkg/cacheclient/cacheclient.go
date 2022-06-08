package cacheclient

import (
	"github.com/QuangTung97/go-memcache/memcache"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"sync/atomic"
	"time"
)

// Client ...
type Client struct {
	client *memcache.Client

	accessCount uint64
	missCount   uint64
}

// Pipeline ...
type Pipeline struct {
	client *Client
	pipe   *memcache.Pipeline

	accessCount uint64
	missCount   uint64
}

var _ dhash.CacheClient = &Client{}

var _ dhash.CachePipeline = &Pipeline{}

// New ...
func New(addr string, numConns int) *Client {
	client, err := memcache.New(addr, numConns, memcache.WithRetryDuration(10*time.Second))
	if err != nil {
		panic(err)
	}
	return &Client{
		client: client,
	}
}

// UnsafeFlushAll ...
func (c *Client) UnsafeFlushAll() error {
	return c.client.Pipeline().FlushAll()()
}

// Close ...
func (c *Client) Close() error {
	return c.client.Close()
}

// Pipeline ...
func (c *Client) Pipeline() dhash.CachePipeline {
	return &Pipeline{
		client: c,
		pipe:   c.client.Pipeline(),
	}
}

// AccessCount ...
func (c *Client) AccessCount() uint64 {
	return atomic.LoadUint64(&c.accessCount)
}

// MissCount ...
func (c *Client) MissCount() uint64 {
	return atomic.LoadUint64(&c.missCount)
}

// Get ...
func (p *Pipeline) Get(key string) func() (dhash.GetOutput, error) {
	fn := p.pipe.MGet(key, memcache.MGetOptions{})
	return func() (dhash.GetOutput, error) {
		p.accessCount++

		resp, err := fn()
		if err != nil {
			return dhash.GetOutput{}, err
		}
		if resp.Type == memcache.MGetResponseTypeVA {
			return dhash.GetOutput{
				Found: true,
				Data:  resp.Data,
			}, nil
		}
		p.missCount++
		return dhash.GetOutput{}, nil
	}
}

// LeaseGet ...
func (p *Pipeline) LeaseGet(key string) func() (dhash.LeaseGetOutput, error) {
	fn := p.pipe.MGet(key, memcache.MGetOptions{
		N:   5,
		CAS: true,
	})
	return func() (dhash.LeaseGetOutput, error) {
		p.accessCount++

		resp, err := fn()
		if err != nil {
			return dhash.LeaseGetOutput{}, err
		}
		if resp.Type != memcache.MGetResponseTypeVA || resp.Flags&memcache.MGetFlagZ != 0 {
			p.missCount++
			return dhash.LeaseGetOutput{
				Type: dhash.LeaseGetTypeRejected,
			}, nil
		}

		if resp.Flags&memcache.MGetFlagW != 0 {
			p.missCount++
			return dhash.LeaseGetOutput{
				Type:    dhash.LeaseGetTypeGranted,
				LeaseID: resp.CAS,
			}, nil
		}

		return dhash.LeaseGetOutput{
			Type: dhash.LeaseGetTypeOK,
			Data: resp.Data,
		}, nil
	}
}

// LeaseSet ...
func (p *Pipeline) LeaseSet(key string, value []byte, leaseID uint64, ttl uint32) func() error {
	fn := p.pipe.MSet(key, value, memcache.MSetOptions{
		CAS: leaseID,
		TTL: ttl,
	})
	return func() error {
		_, err := fn()
		return err
	}
}

// Delete ...
func (p *Pipeline) Delete(key string) func() error {
	fn := p.pipe.MDel(key, memcache.MDelOptions{})
	return func() error {
		_, err := fn()
		return err
	}
}

// Finish ...
func (p *Pipeline) Finish() {
	atomic.AddUint64(&p.client.accessCount, p.accessCount)
	p.accessCount = 0

	atomic.AddUint64(&p.client.missCount, p.missCount)
	p.missCount = 0

	p.pipe.Finish()
}
