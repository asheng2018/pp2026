package payout

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/settlement/model"
)

type PayoutService struct{}

func NewPayoutService() *PayoutService {
	return &PayoutService{}
}

func (s *PayoutService) Execute(ctx context.Context, req *model.PayoutRequest) (string, error) {
	switch req.Method {
	case "USDT":
		return s.payoutUSDT(ctx, req)
	case "wire":
		return s.payoutWire(ctx, req)
	case "wise":
		return s.payoutWise(ctx, req)
	default:
		return "", fmt.Errorf("unsupported payout method: %s", req.Method)
	}
}

func (s *PayoutService) payoutUSDT(ctx context.Context, req *model.PayoutRequest) (string, error) {
	// TRC20 USDT transfer
	log.Info().
		Str("address", maskString(req.Address, 8)).
		Str("amount", req.Amount.String()).
		Msg("executing USDT payout")
	// TODO: Integrate with TRON/Polygon USDT transfer
	txID := fmt.Sprintf("USDT_%s", req.Reference)
	log.Info().Str("tx_id", txID).Msg("USDT payout completed")
	return txID, nil
}

func (s *PayoutService) payoutWire(ctx context.Context, req *model.PayoutRequest) (string, error) {
	log.Info().Str("amount", req.Amount.String()).Msg("executing wire transfer")
	// TODO: Integrate with banking API
	txID := fmt.Sprintf("WIRE_%s", req.Reference)
	return txID, nil
}

func (s *PayoutService) payoutWise(ctx context.Context, req *model.PayoutRequest) (string, error) {
	log.Info().Str("amount", req.Amount.String()).Msg("executing Wise transfer")
	// TODO: Integrate with Wise API
	txID := fmt.Sprintf("WISE_%s", req.Reference)
	return txID, nil
}

func maskString(s string, keep int) string {
	if len(s) <= keep*2 {
		return "***"
	}
	return s[:keep] + "..." + s[len(s)-keep:]
}
