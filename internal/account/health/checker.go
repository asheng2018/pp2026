package health

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/account/model"
)

type HealthChecker struct {
	rdb          *redis.Client
	autoRecover  bool
	maxFails     int
	coolDuration time.Duration
}

func New(rdb *redis.Client, autoRecover bool, maxFails int, coolDuration time.Duration) *HealthChecker {
	if maxFails <= 0 {
		maxFails = 3
	}
	if coolDuration <= 0 {
		coolDuration = 30 * time.Minute
	}
	return &HealthChecker{
		rdb:          rdb,
		autoRecover:  autoRecover,
		maxFails:     maxFails,
		coolDuration: coolDuration,
	}
}

func (hc *HealthChecker) Check(ctx context.Context, accounts []*model.Account) {
	for _, a := range accounts {
		if a.Runtime == nil {
			continue
		}
		if a.Runtime.ConsecutiveFails >= hc.maxFails {
			log.Warn().
				Str("account", a.ID).
				Int("fails", a.Runtime.ConsecutiveFails).
				Msg("account entering cooling state")

			coolUntil := time.Now().Add(hc.coolDuration)
			a.Runtime.CoolingUntil = &coolUntil
			a.Runtime.Status = "cooling"

			key := "account:state:" + a.ID
			hc.rdb.HSet(ctx, key, "status", "cooling", "cooling_until", coolUntil.Format(time.RFC3339))
		}
	}
}

func (hc *HealthChecker) ShouldAutoRecover(a *model.Account) bool {
	if !hc.autoRecover || a.Runtime == nil || a.Runtime.CoolingUntil == nil {
		return false
	}
	return time.Now().After(*a.Runtime.CoolingUntil) && a.Runtime.Status == "cooling"
}

func (hc *HealthChecker) AutoRecover(ctx context.Context, a *model.Account) error {
	log.Info().Str("account", a.ID).Msg("auto-recovering account from cooling")
	a.Runtime.Status = "online"
	a.Runtime.ConsecutiveFails = 0
	a.Runtime.CoolingUntil = nil

	key := "account:state:" + a.ID
	return hc.rdb.HSet(ctx, key, "status", "online", "cooling_until", "").Err()
}
