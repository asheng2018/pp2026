package promotion

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type PromotionCriteria struct {
	MinDaysActive     int
	MinTransactions   int
	MinSuccessRate    float64
	MaxChargebackRate float64
	MinTotalVolume    float64
}

type StagePromoter struct {
	defaultCriteria PromotionCriteria
}

func NewStagePromoter() *StagePromoter {
	return &StagePromoter{
		defaultCriteria: PromotionCriteria{
			MinDaysActive:     7,
			MinTransactions:   20,
			MinSuccessRate:    0.95,
			MaxChargebackRate: 0.01,
			MinTotalVolume:    500.00,
		},
	}
}

func (p *StagePromoter) EvaluateForPromotion(ctx context.Context, accountID, currentStage string, stats AccountStats) (bool, string) {
	criteria := p.defaultCriteria

	if stats.DaysActive < criteria.MinDaysActive {
		return false, "insufficient_days_active"
	}
	if stats.TotalTransactions < criteria.MinTransactions {
		return false, "insufficient_transactions"
	}
	if stats.SuccessRate < criteria.MinSuccessRate {
		return false, "low_success_rate"
	}
	if stats.ChargebackRate > criteria.MaxChargebackRate {
		return false, "high_chargeback_rate"
	}
	if stats.TotalVolume < criteria.MinTotalVolume {
		return false, "insufficient_volume"
	}

	return true, "criteria_met"
}

func (p *StagePromoter) Promote(ctx context.Context, accountID, fromStage, toStage string) error {
	log.Info().
		Str("account", accountID).
		Str("from", fromStage).
		Str("to", toStage).
		Msg("account promoted")
	return nil
}

func (p *StagePromoter) Retire(ctx context.Context, accountID, reason string) error {
	log.Warn().Str("account", accountID).Str("reason", reason).Msg("account retired")
	return nil
}

type AccountStats struct {
	DaysActive        int
	TotalTransactions int
	SuccessRate       float64
	ChargebackRate    float64
	TotalVolume       float64
	LastActive        time.Time
}
