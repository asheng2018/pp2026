package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/account/health"
	"github.com/ab-payment-system/internal/account/model"
	"github.com/ab-payment-system/internal/account/repository"
	"github.com/ab-payment-system/internal/account/state"
)

// CryptoService interface for credential encryption/decryption
type CryptoService interface {
	DecryptString(c string) (string, error)
	EncryptString(p string) (string, error)
}

type AccountService struct {
	repo    *repository.AccountRepo
	state   *state.StateManager
	health  *health.HealthChecker
	crypto  CryptoService
}

func NewAccountService(
	repo *repository.AccountRepo,
	state *state.StateManager,
	health *health.HealthChecker,
	crypto CryptoService,
) *AccountService {
	return &AccountService{
		repo:   repo,
		state:  state,
		health: health,
		crypto: crypto,
	}
}

func (s *AccountService) GetOnlineAccounts(ctx context.Context, gateway, merchantID string) ([]*model.Account, error) {
	accounts, err := s.repo.FindOnline(ctx, gateway, merchantID)
	if err != nil {
		return nil, fmt.Errorf("find online accounts: %w", err)
	}

	for _, a := range accounts {
		rt, _ := s.state.GetRuntime(ctx, a.ID)
		a.Runtime = rt

		// Decrypt credentials into memory
		if len(a.EncryptedCred) > 0 {
			secret, err := s.crypto.DecryptString(string(a.EncryptedCred))
			if err == nil {
				var cred model.Credential
				if json.Unmarshal([]byte(secret), &cred) == nil {
					a.Credential = &cred
				}
			}
		}

		// Auto-recover from cooling
		if s.health.ShouldAutoRecover(a) {
			if err := s.health.AutoRecover(ctx, a); err != nil {
				continue
			}
		}
	}

	// Run health checks
	s.health.Check(ctx, accounts)

	// Filter to only online accounts
	var online []*model.Account
	for _, a := range accounts {
		if a.Runtime != nil && a.Runtime.Status == "online" {
			online = append(online, a)
		}
	}

	return online, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*model.Account, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *AccountService) ReserveAmount(ctx context.Context, id string, amount decimal.Decimal) error {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find account: %w", err)
	}

	ok, msg, err := s.state.ReserveAmount(ctx, id, amount, a.LimitConfig.DailyMax, a.LimitConfig.DailyCount)
	if err != nil {
		return fmt.Errorf("reserve amount: %w", err)
	}
	if !ok {
		return fmt.Errorf("reserve denied: %s", msg)
	}
	return nil
}

func (s *AccountService) ReleaseAmount(ctx context.Context, id string, amount decimal.Decimal) error {
	// Release reservation in Redis
	amtKey := fmt.Sprintf("account:daily:amount:%s", id)
	return s.state.ReleaseAmount(ctx, amtKey, amount)
}

func (s *AccountService) MarkSuccess(ctx context.Context, id string) error {
	return s.state.MarkSuccess(ctx, id)
}

func (s *AccountService) MarkFailure(ctx context.Context, id string) error {
	return s.state.MarkFailure(ctx, id)
}

func (s *AccountService) BatchImport(ctx context.Context, accounts []*model.Account) (int, error) {
	return s.repo.BatchImport(ctx, accounts)
}

func (s *AccountService) CreateAccount(ctx context.Context, a *model.Account) error {
	return s.repo.Create(ctx, a)
}

func (s *AccountService) UpdateStatus(ctx context.Context, id, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}
	return s.state.SetStatus(ctx, id, status)
}

func (s *AccountService) UpdateWeight(ctx context.Context, id string, weight int) error {
	return s.repo.UpdateWeight(ctx, id, weight)
}
