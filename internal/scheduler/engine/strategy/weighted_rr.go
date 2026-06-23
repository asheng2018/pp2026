package strategy

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var ErrNoAccounts = errors.New("no accounts available in strategy")

type WeightedRoundRobin struct {
	counter atomic.Uint64
}

func (s *WeightedRoundRobin) Name() string { return "weighted_round_robin" }

func (s *WeightedRoundRobin) Select(ctx context.Context, accounts []*AccountInfo, _ map[string]string) (*AccountInfo, error) {
	if len(accounts) == 0 {
		return nil, ErrNoAccounts
	}
	totalW := 0
	for _, a := range accounts {
		totalW += a.Weight
	}
	if totalW == 0 {
		// All accounts have zero weight — return the first one
		return accounts[0], nil
	}
	pos := int(s.counter.Add(1) % uint64(totalW))
	for _, a := range accounts {
		pos -= a.Weight
		if pos < 0 {
			log.Ctx(ctx).Debug().Str("account", a.ID).Msg("WRR selected")
			return a, nil
		}
	}
	return accounts[0], nil
}
