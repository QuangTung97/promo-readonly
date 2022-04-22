package readonly

import (
	"context"
	"github.com/shopspring/decimal"
	"time"
)

// IService ...
type IService interface {
	Check(ctx context.Context, inputs []Input) []Output
}

// Input ...
type Input struct {
	ReqTime      time.Time
	VoucherCode  string
	BankCode     string
	Phone        string
	Amount       decimal.Decimal
	MerchantCode string
	TerminalCode string
}

// Output ...
type Output struct {
	DiscountAmount decimal.Decimal
	Err            error
}
