package readonly

import (
	"context"
	"errors"
	"github.com/QuangTung97/promo-readonly/model"
	"github.com/QuangTung97/promo-readonly/repository"
	"github.com/shopspring/decimal"
	"time"
)

//go:generate otelwrap --out service_wrappers.go . IService

// IService ...
type IService interface {
	Check(ctx context.Context, inputs []Input) []Output
}

// Input ...
type Input struct {
	ReqTime      time.Time
	VoucherCode  string
	MerchantCode string
	TerminalCode string
	Phone        string
	BankCode     string
	Amount       decimal.Decimal
}

// Output ...
type Output struct {
	DiscountAmount decimal.Decimal
	Err            error
}

// Service ...
type Service struct {
	provider     repository.Provider
	repoProvider IRepositoryProvider
}

// ErrMerchantInBlacklist ...
var ErrMerchantInBlacklist = errors.New("merchant in blacklist")

// ErrCustomerInBlacklist ...
var ErrCustomerInBlacklist = errors.New("customer in blacklist")

// NewService ...
func NewService(
	provider repository.Provider, repoProvider IRepositoryProvider,
) *Service {
	return &Service{
		provider:     provider,
		repoProvider: repoProvider,
	}
}

type checkState struct {
	repo  IRepository
	ctx   context.Context
	input Input

	getBlacklistMerchant func() (model.NullBlacklistMerchant, error)
	getBlacklistCustomer func() (model.NullBlacklistCustomer, error)

	err error
}

func (s *checkState) setError(err error) {
	s.err = err
}

func (s *checkState) doNext(fn func()) {
	if s.err != nil {
		return
	}
	fn()
}

func (s *checkState) fetchBlacklistMerchant() {
	s.getBlacklistMerchant = s.repo.GetBlacklistMerchant(s.ctx, s.input.MerchantCode)
}

func (s *checkState) fetchBlacklistCustomer() {
	s.getBlacklistCustomer = s.repo.GetBlacklistCustomer(s.ctx, s.input.Phone)
}

func (s *checkState) handleBlacklistMerchant() {
	nullMerchant, err := s.getBlacklistMerchant()
	if err != nil {
		s.setError(err)
		return
	}

	if !nullMerchant.Valid {
		return
	}

	s.setError(ErrMerchantInBlacklist)
}

func (s *checkState) handleBlacklistCustomer() {
	nullCustomer, err := s.getBlacklistCustomer()
	if err != nil {
		s.setError(err)
		return
	}

	if !nullCustomer.Valid {
		return
	}

	s.setError(ErrCustomerInBlacklist)
}

// Check ...
func (s *Service) Check(ctx context.Context, inputs []Input) []Output {
	ctx = s.provider.Readonly(ctx)
	repo := s.repoProvider.NewRepo()
	defer repo.Finish()

	states := make([]*checkState, 0, len(inputs))
	for _, input := range inputs {
		states = append(states, &checkState{
			repo:  repo,
			ctx:   ctx,
			input: input,
		})
	}

	for _, state := range states {
		state.doNext(state.fetchBlacklistMerchant)
		state.doNext(state.fetchBlacklistCustomer)
	}

	for _, state := range states {
		state.doNext(state.handleBlacklistMerchant)
		state.doNext(state.handleBlacklistCustomer)
	}

	outputs := make([]Output, 0, len(states))
	for _, state := range states {
		outputs = append(outputs, Output{
			Err: state.err,
		})
	}
	return outputs
}
