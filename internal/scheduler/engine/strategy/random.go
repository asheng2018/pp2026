package strategy

import (
	"context"
	"math/rand"
)

type RandomSelect struct{}

func (s *RandomSelect) Name() string { return "random" }

func (s *RandomSelect) Select(ctx context.Context, accounts []*AccountInfo, _ map[string]string) (*AccountInfo, error) {
	if len(accounts) == 0 {
		return nil, ErrNoAccounts
	}
	return accounts[rand.Intn(len(accounts))], nil
}
