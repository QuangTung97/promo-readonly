package cacheclient

import (
	"fmt"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func truncateMemcached(c *Client) {
	p := c.client.Pipeline()
	defer p.Finish()
	err := p.FlushAll()()
	if err != nil {
		panic(err)
	}
}

func TestCacheClient__LeaseGet__Granted_And_Rejected(t *testing.T) {
	c := New("localhost:11211", 1)
	truncateMemcached(c)

	p := c.Pipeline()

	output, err := p.LeaseGet("key01")()
	assert.Equal(t, nil, err)

	fmt.Println("LeaseID:", output.LeaseID)

	assert.Greater(t, output.LeaseID, uint64(0))
	output.LeaseID = 0

	assert.Equal(t, dhash.LeaseGetOutput{
		Type: dhash.LeaseGetTypeGranted,
	}, output)

	// Lease Get Second Time
	output, err = p.LeaseGet("key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.LeaseGetOutput{
		Type: dhash.LeaseGetTypeRejected,
	}, output)
}

func TestCacheClient__LeaseGet__OK(t *testing.T) {
	c := New("localhost:11211", 1)
	truncateMemcached(c)

	p := c.Pipeline()

	output, err := p.LeaseGet("key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.LeaseGetTypeGranted, output.Type)

	err = p.LeaseSet("key01", []byte("some value"), output.LeaseID, 0)()
	assert.Equal(t, nil, err)

	// Lease Get After Set
	output, err = p.LeaseGet("key01")()
	assert.Equal(t, nil, err)

	assert.Equal(t, dhash.LeaseGetOutput{
		Type: dhash.LeaseGetTypeOK,
		Data: []byte("some value"),
	}, output)

	// Get
	getOutput, err := p.Get("key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.GetOutput{
		Found: true,
		Data:  []byte("some value"),
	}, getOutput)

	// Get Not Found
	getOutput, err = p.Get("key02")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.GetOutput{}, getOutput)

	// Delete
	err = p.Delete("key01")()
	assert.Equal(t, nil, err)

	getOutput, err = p.Get("key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.GetOutput{}, getOutput)

	// Lease Get Again
	output, err = p.LeaseGet("key01")()
	assert.Equal(t, nil, err)
	assert.Equal(t, dhash.LeaseGetTypeGranted, output.Type)
}
