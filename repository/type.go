package repository

import (
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
)

// HashRange ...
type HashRange struct {
	Begin uint32
	End   dhash.NullUint32
}
