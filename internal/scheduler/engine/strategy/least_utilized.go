package strategy

import (
	"context"
	"sort"
)

type LeastUtilized struct{}

func (s *LeastUtilized) Name() string { return "least_utilized" }

func (s *LeastUtilized) Select(ctx context.Context, accounts []*AccountInfo, _ map[string]string) (*AccountInfo, error) {
	if len(accounts) == 0 {
		return nil, ErrNoAccounts
	}
	sort.Slice(accounts, func(i, j int) bool {
		ri := accounts[i].TodayAmount.InexactFloat64() / accounts[i].DailyMax.InexactFloat64()
		rj := accounts[j].TodayAmount.InexactFloat64() / accounts[j].DailyMax.InexactFloat64()
		return ri < rj
	})
	return accounts[0], nil
}
