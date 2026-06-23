package strategy

import (
	"context"
	"sync"
)

type Sequential struct {
	mu      sync.Mutex
	lastIdx map[string]int
}

func NewSequential() *Sequential {
	return &Sequential{lastIdx: make(map[string]int)}
}

func (s *Sequential) Name() string { return "sequential" }

func (s *Sequential) Select(ctx context.Context, accounts []*AccountInfo, metadata map[string]string) (*AccountInfo, error) {
	if len(accounts) == 0 {
		return nil, ErrNoAccounts
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	group := metadata["merchant_id"]
	idx := s.lastIdx[group]
	next := (idx + 1) % len(accounts)
	s.lastIdx[group] = next
	return accounts[next], nil
}
