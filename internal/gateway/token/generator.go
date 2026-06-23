package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid payment token")
	ErrExpiredToken  = errors.New("payment token expired")
)

type PayToken struct {
	TokenID   string    `json:"tid"`
	AccountID string    `json:"aid"`
	OrderID   string    `json:"oid"`
	Amount    string    `json:"amt"`
	Gateway   string    `json:"gw"`
	ExpiresAt time.Time `json:"exp"`
	CreatedAt time.Time `json:"iat"`
}

type Generator struct {
	secretKey []byte
}

func NewGenerator(secret string) *Generator {
	return &Generator{secretKey: []byte(secret)}
}

func (g *Generator) Generate(accountID, orderID, amount, gateway string, ttl time.Duration) (string, error) {
	now := time.Now()
	payload := &PayToken{
		TokenID:   uuid.New().String(),
		AccountID: accountID,
		OrderID:   orderID,
		Amount:    amount,
		Gateway:   gateway,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(data)
	mac := hmac.New(sha256.New, g.secretKey)
	mac.Write([]byte(encoded))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return encoded + "." + sig, nil
}

func (g *Generator) Validate(tokenStr string) (*PayToken, error) {
	parts := strings.SplitN(tokenStr, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, g.secretKey)
	mac.Write([]byte(parts[0]))
	expectedSig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}
	data, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var token PayToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrExpiredToken
	}
	return &token, nil
}
