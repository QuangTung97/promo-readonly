package util

import "github.com/twmb/murmur3"

// HashFunc ...
func HashFunc(s string) uint32 {
	return murmur3.Sum32([]byte(s))
}
