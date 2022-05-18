package util

import "github.com/spaolacci/murmur3"

// HashFunc ...
func HashFunc(s string) uint32 {
	return murmur3.Sum32([]byte(s))
}
