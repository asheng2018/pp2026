package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PaymentSession struct {
	TokenID    string    `json:"token_id"`
	OrderID    string    `json:"order_id"`
	AccountID  string    `json:"account_id"`
	Amount     string    `json:"amount"`
	Gateway    string    `json:"gateway"`
	Status     string    `json:"status"`
	MerchantID string    `json:"merchant_id"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type SessionManager struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewSessionManager(rdb *redis.Client, ttl time.Duration) *SessionManager {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &SessionManager{rdb: rdb, ttl: ttl}
}

func (sm *SessionManager) key(tokenID string) string {
	return fmt.Sprintf("pay:session:%s", tokenID)
}

func (sm *SessionManager) Create(ctx context.Context, session *PaymentSession) error {
	session.CreatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(sm.ttl)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return sm.rdb.Set(ctx, sm.key(session.TokenID), data, sm.ttl).Err()
}

func (sm *SessionManager) Get(ctx context.Context, tokenID string) (*PaymentSession, error) {
	data, err := sm.rdb.Get(ctx, sm.key(tokenID)).Bytes()
	if err != nil {
		return nil, err
	}
	var session PaymentSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (sm *SessionManager) UpdateStatus(ctx context.Context, tokenID, status string) error {
	session, err := sm.Get(ctx, tokenID)
	if err != nil {
		return err
	}
	session.Status = status
	data, _ := json.Marshal(session)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute
	}
	return sm.rdb.Set(ctx, sm.key(tokenID), data, ttl).Err()
}

func (sm *SessionManager) Delete(ctx context.Context, tokenID string) error {
	return sm.rdb.Del(ctx, sm.key(tokenID)).Err()
}
