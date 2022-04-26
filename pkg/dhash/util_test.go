package dhash

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalUnmarshal_Single_Entry(t *testing.T) {
	data := marshalEntries([]Entry{
		{
			Hash: 55,
			Data: []byte{10, 12, 14},
		},
	})
	assert.Equal(t, []byte{
		1,           // number of items
		55, 0, 0, 0, // hash
		3,          // data size
		10, 12, 14, // data
	}, data)

	entries, err := unmarshalEntries(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, []Entry{
		{
			Hash: 55,
			Data: []byte{10, 12, 14},
		},
	}, entries)
}

func repeatBytes(a byte, n int) []byte {
	result := make([]byte, n)
	for i := range result {
		result[i] = a
	}
	return result
}

func TestMarshalUnmarshal_Multiple_Entries(t *testing.T) {
	assert.Equal(t, []byte{8, 8}, repeatBytes(8, 2))

	entries := []Entry{
		{
			Hash: 55,
			Data: []byte{10, 12, 14},
		},
		{
			Hash: 80,
			Data: []byte{30, 31, 32, 33, 34, 35},
		},
		{
			Hash: 0x778899aa,
			Data: repeatBytes(0x9, 345),
		},
		{
			Hash: 0x664542aa,
			Data: []byte{99, 99, 88, 88},
		},
	}

	data := marshalEntries(entries)

	results, err := unmarshalEntries(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, entries, results)
}

func TestUnmarshal_Error__Missing_Entry_Count(t *testing.T) {
	_, err := unmarshalEntries(nil)
	assert.Equal(t, errors.New("unmarshal entries: invalid entry count"), err)
}

func TestUnmarshal_Error__Missing_Hash(t *testing.T) {
	_, err := unmarshalEntries([]byte{1, 0x5, 0x6, 0x7})
	assert.Equal(t, errors.New("unmarshal entries: missing bytes for hash"), err)
}

func TestUnmarshal_Error__Missing_Data_Len(t *testing.T) {
	_, err := unmarshalEntries([]byte{1, 0x5, 0x6, 0x7, 0x8})
	assert.Equal(t, errors.New("unmarshal entries: missing data length"), err)
}

func TestUnmarshal_Error__Missing_Data(t *testing.T) {
	_, err := unmarshalEntries([]byte{1, 0x5, 0x6, 0x7, 0x8, 3, 0xa, 0xb})
	assert.Equal(t, errors.New("unmarshal entries: missing data"), err)
}
