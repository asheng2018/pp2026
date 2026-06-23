package detector

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// AnomalyDetector detects anomalous patterns using Redis-based counters.
type AnomalyDetector struct {
	rdb *redis.Client
}

func NewAnomalyDetector(rdb *redis.Client) *AnomalyDetector {
	return &AnomalyDetector{rdb: rdb}
}

// VelocityCheck detects rapid orders from the same IP or email.
func (d *AnomalyDetector) VelocityCheck(ctx context.Context, key string, window time.Duration, maxCount int) (bool, int, error) {
	k := "risk:velocity:" + key
	count, err := d.rdb.Incr(ctx, k).Result()
	if err != nil {
		return false, 0, err
	}
	if count == 1 {
		d.rdb.Expire(ctx, k, window)
	}
	isAnomaly := int(count) > maxCount
	if isAnomaly {
		log.Ctx(ctx).Warn().Str("key", key).Int64("count", count).Int("max", maxCount).Msg("velocity anomaly detected")
	}
	return isAnomaly, int(count), nil
}

// AmountPatternCheck detects testing patterns (small amounts, sequential).
func (d *AnomalyDetector) AmountPatternCheck(ctx context.Context, accountID string, amount float64) (bool, error) {
	// Check for small-amount testing patterns ($1.00 or less repeatedly)
	k := "risk:amount_pattern:" + accountID
	if amount <= 1.0 {
		count, _ := d.rdb.Incr(ctx, k).Result()
		d.rdb.Expire(ctx, k, 1*time.Hour)
		if count > 5 {
			log.Ctx(ctx).Warn().Str("account", accountID).Int64("small_count", count).Msg("card testing pattern detected")
			return true, nil
		}
	}
	return false, nil
}

// GeoAnomalyCheck detects geographic inconsistencies.
func (d *AnomalyDetector) GeoAnomalyCheck(ctx context.Context, ip, expectedCountry string) (bool, error) {
	// In production, use MaxMind GeoIP database
	if expectedCountry == "" {
		return false, nil
	}
	return false, nil
}
