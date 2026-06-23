package strategy

import (
	"context"

	"github.com/shopspring/decimal"
)

type AccountInfo struct {
	ID                string
	Gateway           string
	Status            string
	Weight            int
	Priority          int
	TodayAmount       decimal.Decimal
	DailyMax          decimal.Decimal
	SuccessRate       float64
	ConsecutiveFails  int
	LastUsedAt        int64
}

type RoutingStrategy interface {
	Select(ctx context.Context, accounts []*AccountInfo, metadata map[string]string) (*AccountInfo, error)
	Name() string
}
