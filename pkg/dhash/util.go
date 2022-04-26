package dhash

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func unmarshalError(msg string) error {
	return fmt.Errorf("unmarshal entries: %s", msg)
}

func unmarshalEntries(data []byte) ([]Entry, error) {
	count, n := binary.Uvarint(data)
	if n <= 0 {
		return nil, unmarshalError("invalid entry count")
	}

	data = data[n:]

	results := make([]Entry, count)
	for i := range results {
		// hash
		if len(data) < 4 {
			return nil, unmarshalError("missing bytes for hash")
		}
		hash := binary.LittleEndian.Uint32(data)
		data = data[4:]

		// data length
		dataLen, n := binary.Uvarint(data)
		if n <= 0 {
			return nil, unmarshalError("missing data length")
		}
		data = data[n:]

		// data
		if uint64(len(data)) < dataLen {
			return nil, unmarshalError("missing data")
		}
		d := make([]byte, dataLen)
		copy(d, data)
		data = data[dataLen:]

		results[i] = Entry{
			Hash: hash,
			Data: d,
		}
	}

	return results, nil
}

func marshalEntries(entries []Entry) []byte {
	var buf bytes.Buffer
	var placeholder [16]byte

	// number of items
	size := binary.PutUvarint(placeholder[:], uint64(len(entries)))
	_, _ = buf.Write(placeholder[:size])

	for _, entry := range entries {
		// hash
		binary.LittleEndian.PutUint32(placeholder[:], entry.Hash)
		_, _ = buf.Write(placeholder[:4])

		// data len
		size := binary.PutUvarint(placeholder[:], uint64(len(entry.Data)))
		_, _ = buf.Write(placeholder[:size])

		_, _ = buf.Write(entry.Data)
	}

	return buf.Bytes()
}

func startOfSlot(hash uint32, sizeLog int) uint32 {
	mask := uint32(0xffffffff << (32 - sizeLog))
	return hash & mask
}
