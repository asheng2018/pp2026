package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/account/model"
	"github.com/ab-payment-system/internal/scheduler/circuit"
	"github.com/ab-payment-system/internal/scheduler/dto"
	"github.com/ab-payment-system/internal/scheduler/engine/strategy"
	pkgerrors "github.com/ab-payment-system/pkg/errors"
)

// AccountPool is the interface for account operations needed by the orchestrator.
type AccountPool interface {
	GetOnlineAccounts(ctx context.Context, gateway string, merchantID string) ([]*model.Account, error)
	GetAccount(ctx context.Context, id string) (*model.Account, error)
	ReserveAmount(ctx context.Context, id string, amount decimal.Decimal) error
	ReleaseAmount(ctx context.Context, id string, amount decimal.Decimal) error
	MarkSuccess(ctx context.Context, id string) error
	MarkFailure(ctx context.Context, id string) error
}

type Orchestrator struct {
	accountPool       AccountPool
	strategies        map[string]strategy.RoutingStrategy
	filterChain       *FilterChain
	breakers          map[string]*circuit.Breaker
	breakerThreshold  int
	breakerTimeout    time.Duration
}

func NewOrchestrator(pool AccountPool) *Orchestrator {
	o := &Orchestrator{
		accountPool:      pool,
		breakers:         make(map[string]*circuit.Breaker),
		breakerThreshold: 3,
		breakerTimeout:   60 * time.Second,
		strategies: map[string]strategy.RoutingStrategy{
			"weighted_round_robin": &strategy.WeightedRoundRobin{},
			"sequential":           strategy.NewSequential(),
			"random":               &strategy.RandomSelect{},
			"least_utilized":       &strategy.LeastUtilized{},
		},
	}
	o.filterChain = NewFilterChain(
		&StatusFilter{},
		&AmountFilter{},
		&DailyLimitFilter{},
		&ConsecutiveFailFilter{MaxFails: 5},
		&SuccessRateFilter{MinRate: 0.7},
	)
	return o
}

func (o *Orchestrator) Allocate(ctx context.Context, req *dto.AllocateRequest) (*dto.AllocateResponse, error) {
	accounts, err := o.accountPool.GetOnlineAccounts(ctx, req.Gateway, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("get online accounts: %w", err)
	}

	if len(accounts) == 0 {
		return nil, pkgerrors.ErrNoAccountAvailable
	}

	// Apply filter chain
	fc := FilterContext{
		Amount:     req.Amount.String(),
		Currency:   req.Currency,
		Country:    req.Country,
		MerchantID: req.MerchantID,
	}
	candidates := o.filterChain.Apply(ctx, accounts, fc)

	if len(candidates) == 0 {
		return nil, pkgerrors.ErrNoAccountAvailable
	}

	// Convert to strategy AccountInfo
	var stratInfo []*strategy.AccountInfo
	for _, a := range candidates {
		info := &strategy.AccountInfo{
			ID:       a.ID,
			Gateway:  string(a.Gateway),
			Status:   string(a.Status),
			Weight:   a.Weight,
			Priority: a.Priority,
		}
		if a.Runtime != nil {
			info.TodayAmount = a.Runtime.TodayAmount
			info.SuccessRate = a.Runtime.SuccessRate
			info.ConsecutiveFails = a.Runtime.ConsecutiveFails
			info.LastUsedAt = a.Runtime.LastUsedAt.Unix()
		}
		info.DailyMax = a.LimitConfig.DailyMax
		stratInfo = append(stratInfo, info)
	}

	// Select using strategy
	s, ok := o.strategies[req.Strategy]
	if !ok {
		s = o.strategies["weighted_round_robin"]
	}

	selected, err := s.Select(ctx, stratInfo, req.Metadata)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "STRATEGY_ERROR", "failed to select account", 500)
	}

	// Reserve amount
	if err := o.accountPool.ReserveAmount(ctx, selected.ID, req.Amount); err != nil {
		return nil, pkgerrors.Wrap(err, "RESERVE_ERROR", "failed to reserve amount", 503)
	}

	token := uuid.New().String()
	log.Ctx(ctx).Info().
		Str("order_id", req.OrderID).
		Str("account", selected.ID).
		Str("gateway", req.Gateway).
		Str("strategy", s.Name()).
		Msg("account allocated")

	return &dto.AllocateResponse{
		PayToken:   token,
		GatewayURL: fmt.Sprintf("/pay/%s", token),
		AccountRef: selected.ID[:8],
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		Gateway:    req.Gateway,
	}, nil
}

func (o *Orchestrator) Release(ctx context.Context, req *dto.ReleaseRequest) error {
	log.Ctx(ctx).Info().Str("order_id", req.OrderID).Msg("releasing allocation")
	return nil
}

func (o *Orchestrator) Confirm(ctx context.Context, req *dto.ConfirmRequest) error {
	log.Ctx(ctx).Info().Str("order_id", req.OrderID).Msg("confirming payment")
	return nil
}
