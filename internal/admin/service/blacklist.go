package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist manages JWT token revocation via Redis.
type TokenBlacklist struct {
	client *redis.Client
}

// NewTokenBlacklist creates a new TokenBlacklist backed by the given Redis client.
func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// Blacklist stores the given jti (JWT ID) in Redis with a TTL.
// After the TTL expires, the entry auto-deletes and the token can no longer
// be used anyway because it will have expired naturally.
func (b *TokenBlacklist) Blacklist(ctx context.Context, jti string, ttl time.Duration) error {
	return b.client.Set(ctx, "jwt:blacklist:"+jti, "1", ttl).Err()
}

// IsBlacklisted checks whether the given jti has been revoked.
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := b.client.Exists(ctx, "jwt:blacklist:"+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
