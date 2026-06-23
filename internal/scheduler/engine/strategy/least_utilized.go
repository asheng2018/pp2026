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
		di := accounts[i].DailyMax.InexactFloat64()
		dj := accounts[j].DailyMax.InexactFloat64()
		if di == 0 {
			di = 1
		}
		if dj == 0 {
			dj = 1
		}
		ri := accounts[i].TodayAmount.InexactFloat64() / di
		rj := accounts[j].TodayAmount.InexactFloat64() / dj
		return ri < rj
	})
	return accounts[0], nil
}
