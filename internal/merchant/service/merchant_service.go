package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/merchant/model"
)

type MerchantRepository interface {
	FindByID(ctx context.Context, id string) (*model.Merchant, error)
	FindByAPIKeyHash(ctx context.Context, keyHash string) (*model.Merchant, *model.APIKey, error)
	Create(ctx context.Context, m *model.Merchant) error
	Update(ctx context.Context, m *model.Merchant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*model.Merchant, error)
	CreateAPIKey(ctx context.Context, merchantID string, key *model.APIKey) error
	RevokeAPIKey(ctx context.Context, keyID string) error
	ListAPIKeys(ctx context.Context, merchantID string) ([]*model.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error
}

type MerchantService struct {
	repo      MerchantRepository
	jwtSecret string
}

func NewMerchantService(repo MerchantRepository, jwtSecret string) *MerchantService {
	return &MerchantService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *MerchantService) Create(ctx context.Context, m *model.Merchant) error {
	m.ID = uuid.New().String()
	m.Status = "active"
	if m.RoutingMode == "" {
		m.RoutingMode = "weighted_round_robin"
	}
	if m.FeeConfig == nil {
		m.FeeConfig = &model.FeeConfig{
			PayPalFee: &model.GatewayFee{FixedFee: decimalFromString("0.30"), RateFee: decimalFromString("0.044")},
			StripeFee: &model.GatewayFee{FixedFee: decimalFromString("0.30"), RateFee: decimalFromString("0.039")},
		}
	}
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return s.repo.Create(ctx, m)
}

func (s *MerchantService) Get(ctx context.Context, id string) (*model.Merchant, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *MerchantService) Update(ctx context.Context, m *model.Merchant) error {
	m.UpdatedAt = time.Now()
	return s.repo.Update(ctx, m)
}

func (s *MerchantService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *MerchantService) List(ctx context.Context, offset, limit int) ([]*model.Merchant, error) {
	return s.repo.List(ctx, offset, limit)
}

// GenerateAPIKey creates a new API key for a merchant.
func (s *MerchantService) GenerateAPIKey(ctx context.Context, merchantID string, permissions []string, ipWhitelist []string, expiresAt *time.Time) (*model.APIKey, string, error) {
	rawKey := "ab_" + uuid.New().String() + "_" + hex.EncodeToString(secureRandom(16))
	keyHash := hashKey(rawKey)

	key := &model.APIKey{
		KeyID:       uuid.New().String(),
		KeyHash:     keyHash,
		KeyPrefix:   rawKey[:10],
		Permissions: permissions,
		IPWhitelist: ipWhitelist,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.CreateAPIKey(ctx, merchantID, key); err != nil {
		return nil, "", fmt.Errorf("create api key: %w", err)
	}

	// Return the raw key only once
	return key, rawKey, nil
}

func (s *MerchantService) RevokeAPIKey(ctx context.Context, keyID string) error {
	return s.repo.RevokeAPIKey(ctx, keyID)
}

func (s *MerchantService) ListAPIKeys(ctx context.Context, merchantID string) ([]*model.APIKey, error) {
	return s.repo.ListAPIKeys(ctx, merchantID)
}

// AuthenticateAPIKey validates an API key and returns the merchant.
func (s *MerchantService) AuthenticateAPIKey(ctx context.Context, rawKey string) (*model.Merchant, *model.APIKey, error) {
	keyHash := hashKey(rawKey)

	merchant, key, err := s.repo.FindByAPIKeyHash(ctx, keyHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid api key: %w", err)
	}

	if key.RevokedAt != nil {
		return nil, nil, fmt.Errorf("api key revoked")
	}
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, nil, fmt.Errorf("api key expired")
	}
	if merchant.Status != "active" {
		return nil, nil, fmt.Errorf("merchant %s", merchant.Status)
	}

	// Update last used
	go s.repo.UpdateAPIKeyLastUsed(context.Background(), key.KeyID)

	return merchant, key, nil
}

// VerifySignature validates an HMAC-SHA256 request signature.
func (s *MerchantService) VerifySignature(payload, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 1
}

// CalculateFee computes the fee for a given transaction.
func (s *MerchantService) CalculateFee(merchant *model.Merchant, gateway string, amount decimal.Decimal) decimal.Decimal {
	var fee *model.GatewayFee
	switch gateway {
	case "paypal":
		if merchant.FeeConfig != nil {
			fee = merchant.FeeConfig.PayPalFee
		}
	case "stripe":
		if merchant.FeeConfig != nil {
			fee = merchant.FeeConfig.StripeFee
		}
	}
	if fee == nil {
		return decimal.Zero
	}

	calculated := amount.Mul(fee.RateFee).Add(fee.FixedFee)
	if calculated.LessThan(fee.MinFee) {
		return fee.MinFee
	}
	return calculated
}

func (s *MerchantService) Suspend(ctx context.Context, id string) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	m.Status = "suspended"
	m.UpdatedAt = time.Now()
	return s.repo.Update(ctx, m)
}

func (s *MerchantService) Activate(ctx context.Context, id string) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	m.Status = "active"
	m.UpdatedAt = time.Now()
	return s.repo.Update(ctx, m)
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func secureRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return b
}

func decimalFromString(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
