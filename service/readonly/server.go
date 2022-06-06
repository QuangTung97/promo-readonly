package readonly

import (
	"context"
	"fmt"
	"github.com/QuangTung97/promo-readonly/pkg/dhash"
	"github.com/QuangTung97/promo-readonly/promopb"
	"github.com/QuangTung97/promo-readonly/repository"
	"go.opentelemetry.io/otel"
)

// Server ...
type Server struct {
	promopb.UnimplementedPromoServiceServer
	service IService
}

// NewServer ...
func NewServer(provider repository.Provider, dhashProvider dhash.Provider) *Server {
	blacklistRepo := repository.NewBlacklist()
	s := NewService(provider, blacklistRepo, dhashProvider)
	return &Server{
		service: NewIServiceWrapper(s,
			otel.GetTracerProvider().Tracer("server"), "service::"),
	}
}

// Check ...
func (s *Server) Check(
	ctx context.Context, req *promopb.PromoServiceCheckRequest,
) (*promopb.PromoServiceCheckResponse, error) {
	inputs := make([]Input, 0, len(req.Inputs))
	for _, input := range req.Inputs {
		inputs = append(inputs, Input{
			ReqTime:      req.ReqTime.AsTime(),
			VoucherCode:  input.VoucherCode,
			MerchantCode: input.MerchantCode,
			TerminalCode: input.TerminalCode,
			Phone:        input.Phone,
		})
	}

	outputs := s.service.Check(ctx, inputs)
	// fmt.Println(outputs)

	respOutputs := make([]*promopb.PromoServiceCheckOutput, 0, len(outputs))
	for index, o := range outputs {
		status := int32(1)
		if o.Err != nil {
			if o.Err != ErrCustomerInBlacklist && o.Err != ErrMerchantInBlacklist {
				fmt.Println(index, o.Err)
			}
			status = 2
		}

		respOutputs = append(respOutputs, &promopb.PromoServiceCheckOutput{
			Status: status,
		})
	}

	return &promopb.PromoServiceCheckResponse{
		Outputs: respOutputs,
	}, nil
}
