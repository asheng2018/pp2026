package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/settlement/model"
)

type FeeCalculator interface {
	CalculateFee(merchantID string, gateway string, amount decimal.Decimal) decimal.Decimal
}

type OrderSummary struct {
	TotalOrders int
	TotalAmount decimal.Decimal
	Orders      []OrderDetail
}

type OrderDetail struct {
	OrderID string
	Amount  decimal.Decimal
	Gateway string
	Fee     decimal.Decimal
	PaidAt  time.Time
}

type SettlementRepo interface {
	Create(ctx context.Context, s *model.Settlement) error
	FindByID(ctx context.Context, id string) (*model.Settlement, error)
	FindByMerchant(ctx context.Context, merchantID string, offset, limit int) ([]*model.Settlement, error)
	FindPending(ctx context.Context) ([]*model.Settlement, error)
	UpdateStatus(ctx context.Context, id string, status string, txID string) error
}

type PayoutService interface {
	Execute(ctx context.Context, req *model.PayoutRequest) (string, error) // returns txID
}

type SettlementEngine struct {
	repo        SettlementRepo
	feeCalc     FeeCalculator
	payoutSvc   PayoutService
}

func New(repo SettlementRepo, feeCalc FeeCalculator, payoutSvc PayoutService) *SettlementEngine {
	return &SettlementEngine{
		repo:      repo,
		feeCalc:   feeCalc,
		payoutSvc: payoutSvc,
	}
}

// GenerateSettlement creates a settlement for completed orders in a cycle.
func (e *SettlementEngine) GenerateSettlement(ctx context.Context, merchantID string, cycleStart, cycleEnd time.Time, summary *OrderSummary) (*model.Settlement, error) {
	var totalFee decimal.Decimal
	for _, o := range summary.Orders {
		fee := e.feeCalc.CalculateFee(merchantID, o.Gateway, o.Amount)
		totalFee = totalFee.Add(fee)
	}

	netAmount := summary.TotalAmount.Sub(totalFee)

	settlement := &model.Settlement{
		ID:          uuid.New().String(),
		MerchantID:  merchantID,
		CycleStart:  cycleStart,
		CycleEnd:    cycleEnd,
		TotalOrders: summary.TotalOrders,
		TotalAmount: summary.TotalAmount,
		TotalFee:    totalFee,
		NetAmount:   netAmount,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := e.repo.Create(ctx, settlement); err != nil {
		return nil, fmt.Errorf("create settlement: %w", err)
	}

	log.Info().
		Str("settlement", settlement.ID).
		Str("merchant", merchantID).
		Str("total", summary.TotalAmount.String()).
		Str("fee", totalFee.String()).
		Str("net", netAmount.String()).
		Msg("settlement generated")

	return settlement, nil
}

// ProcessPending processes all pending settlements.
func (e *SettlementEngine) ProcessPending(ctx context.Context) (int, int, error) {
	settlements, err := e.repo.FindPending(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("find pending settlements: %w", err)
	}

	processed := 0
	failed := 0

	for _, s := range settlements {
		req := &model.PayoutRequest{
			SettlementID: s.ID,
			Method:       s.PayoutMethod,
			Address:      s.PayoutAddress,
			Amount:       s.NetAmount,
			Currency:     "USD",
			Reference:    fmt.Sprintf("SETTLE_%s", s.ID[:8]),
		}

		txID, err := e.payoutSvc.Execute(ctx, req)
		if err != nil {
			failed++
			log.Error().Err(err).Str("settlement", s.ID).Msg("payout failed")
			e.repo.UpdateStatus(ctx, s.ID, "failed", "")
			continue
		}

		now := time.Now()
		s.Status = "completed"
		s.PayoutTxID = txID
		s.SettledAt = &now

		if err := e.repo.UpdateStatus(ctx, s.ID, s.Status, txID); err != nil {
			failed++
			log.Error().Err(err).Str("settlement", s.ID).Msg("failed to update settlement status")
			continue
		}
		processed++
		log.Info().Str("settlement", s.ID).Str("tx_id", txID).Msg("settlement completed")
	}

	return processed, failed, nil
}
