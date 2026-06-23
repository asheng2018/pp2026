package engine

import (
	"context"
	"fmt"

	"github.com/ab-payment-system/internal/account/model"
)

type AccountFilter interface {
	Name() string
	Filter(ctx context.Context, account *model.Account, req FilterContext) error
}

type FilterContext struct {
	Amount     string
	Currency   string
	Country    string
	MerchantID string
}

type FilterChain struct {
	filters []AccountFilter
}

func NewFilterChain(filters ...AccountFilter) *FilterChain {
	return &FilterChain{filters: filters}
}

func (fc *FilterChain) Add(f AccountFilter) {
	fc.filters = append(fc.filters, f)
}

func (fc *FilterChain) Apply(ctx context.Context, accounts []*model.Account, req FilterContext) []*model.Account {
	var result []*model.Account
	for _, a := range accounts {
		pass := true
		for _, f := range fc.filters {
			if err := f.Filter(ctx, a, req); err != nil {
				pass = false
				break
			}
		}
		if pass {
			result = append(result, a)
		}
	}
	return result
}

// Built-in Filters

type StatusFilter struct{}

func (f *StatusFilter) Name() string { return "status" }

func (f *StatusFilter) Filter(ctx context.Context, a *model.Account, _ FilterContext) error {
	if a.Status != "online" {
		return fmt.Errorf("account %s status is %s", a.ID, a.Status)
	}
	return nil
}

type AmountFilter struct{}

func (f *AmountFilter) Name() string { return "amount_filter" }

func (f *AmountFilter) Filter(ctx context.Context, a *model.Account, req FilterContext) error {
	return nil
}

type DailyLimitFilter struct{}

func (f *DailyLimitFilter) Name() string { return "daily_limit" }

func (f *DailyLimitFilter) Filter(ctx context.Context, a *model.Account, req FilterContext) error {
	if a.Runtime != nil && a.Runtime.TodayAmount.GreaterThan(a.LimitConfig.DailyMax) {
		return fmt.Errorf("daily limit exceeded for account %s", a.ID)
	}
	return nil
}

type ConsecutiveFailFilter struct{ MaxFails int }

func (f *ConsecutiveFailFilter) Name() string { return "consecutive_fail" }

func (f *ConsecutiveFailFilter) Filter(ctx context.Context, a *model.Account, _ FilterContext) error {
	if a.Runtime != nil && a.Runtime.ConsecutiveFails >= f.MaxFails {
		return fmt.Errorf("circuit breaker: %d consecutive fails", a.Runtime.ConsecutiveFails)
	}
	return nil
}

type SuccessRateFilter struct{ MinRate float64 }

func (f *SuccessRateFilter) Name() string { return "success_rate" }

func (f *SuccessRateFilter) Filter(ctx context.Context, a *model.Account, _ FilterContext) error {
	if a.Runtime != nil && a.Runtime.SuccessRate < f.MinRate {
		return fmt.Errorf("success rate %.2f below min %.2f", a.Runtime.SuccessRate, f.MinRate)
	}
	return nil
}
