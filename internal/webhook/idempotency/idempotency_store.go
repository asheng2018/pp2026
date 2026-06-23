package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewIdempotencyStore(rdb *redis.Client, ttl time.Duration) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &IdempotencyStore{rdb: rdb, ttl: ttl}
}

func (s *IdempotencyStore) IsDuplicate(ctx context.Context, eventID string) (bool, error) {
	key := "idempotency:" + eventID
	ok, err := s.rdb.SetNX(ctx, key, "processed", s.ttl).Result()
	if err != nil {
		return false, err
	}
	// If SetNX returns false, the key already exists — it's a duplicate
	return !ok, nil
}

func (s *IdempotencyStore) MarkProcessed(ctx context.Context, eventID string) error {
	key := "idempotency:" + eventID
	return s.rdb.Set(ctx, key, "processed", s.ttl).Err()
}

func (s *IdempotencyStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	key := "idempotency:" + eventID
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
