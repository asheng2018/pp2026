package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/account/model"
)

const (
	reserveScript = `
local amt_key = KEYS[1]
local cnt_key = KEYS[2]
local amount = tonumber(ARGV[1])
local daily_max = tonumber(ARGV[2])
local daily_count_max = tonumber(ARGV[3])

local current_amt = redis.call('INCRBYFLOAT', amt_key, amount)
local current_cnt = redis.call('INCR', cnt_key)

local now = redis.call('TIME')
local endOfDay = 86400 - (now[1] % 86400)
redis.call('EXPIRE', amt_key, endOfDay)
redis.call('EXPIRE', cnt_key, endOfDay)

if current_amt > daily_max then
  redis.call('INCRBYFLOAT', amt_key, 0 - amount)
  redis.call('DECR', cnt_key)
  return {0, 'daily_amount_exceeded'}
end

if current_cnt > daily_count_max then
  redis.call('INCRBYFLOAT', amt_key, 0 - amount)
  redis.call('DECR', cnt_key)
  return {0, 'daily_count_exceeded'}
end

return {1, 'ok'}
`
)

type StateManager struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewStateManager(rdb *redis.Client) *StateManager {
	return &StateManager{rdb: rdb, ttl: 24 * time.Hour}
}

func (sm *StateManager) key(accountID string) string {
	return fmt.Sprintf("account:state:%s", accountID)
}

func (sm *StateManager) GetRuntime(ctx context.Context, accountID string) (*model.AccountRuntime, error) {
	data, err := sm.rdb.Get(ctx, sm.key(accountID)).Bytes()
	if err != nil {
		return sm.defaultRuntime(accountID), nil
	}
	var rt model.AccountRuntime
	if err := json.Unmarshal(data, &rt); err != nil {
		return sm.defaultRuntime(accountID), nil
	}
	return &rt, nil
}

func (sm *StateManager) SetRuntime(ctx context.Context, rt *model.AccountRuntime) error {
	data, err := json.Marshal(rt)
	if err != nil {
		return err
	}
	return sm.rdb.Set(ctx, sm.key(rt.AccountID), data, sm.ttl).Err()
}

func (sm *StateManager) defaultRuntime(accountID string) *model.AccountRuntime {
	return &model.AccountRuntime{
		AccountID:        accountID,
		Status:           "online",
		TodayAmount:      decimal.Zero,
		SuccessRate:      1.0,
	}
}

// ReserveAmount atomically reserves amount against daily limits
func (sm *StateManager) ReserveAmount(
	ctx context.Context, accountID string,
	amount, dailyMax decimal.Decimal, dailyCountMax int,
) (bool, string, error) {
	amtKey := fmt.Sprintf("account:daily:amount:%s", accountID)
	cntKey := fmt.Sprintf("account:daily:count:%s", accountID)

	res, err := sm.rdb.Eval(ctx, reserveScript,
		[]string{amtKey, cntKey},
		amount.InexactFloat64(),
		dailyMax.InexactFloat64(),
		dailyCountMax,
	).Result()
	if err != nil {
		return false, "", err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, "unexpected_result", fmt.Errorf("unexpected reserve result: %v", res)
	}

	okVal, _ := arr[0].(int64)
	msg, _ := arr[1].(string)
	return okVal == 1, msg, nil
}

// ReleaseAmount decrements the daily amount counter
func (sm *StateManager) ReleaseAmount(ctx context.Context, amtKey string, amount decimal.Decimal) error {
	return sm.rdb.IncrByFloat(ctx, amtKey, 0-amount.InexactFloat64()).Err()
}

func (sm *StateManager) MarkSuccess(ctx context.Context, accountID string) error {
	rt, err := sm.GetRuntime(ctx, accountID)
	if err != nil {
		return err
	}
	rt.ConsecutiveFails = 0
	rt.LastUsedAt = time.Now()
	total := float64(100)
	rt.SuccessRate = (rt.SuccessRate*(total-1) + 1) / total
	return sm.SetRuntime(ctx, rt)
}

func (sm *StateManager) MarkFailure(ctx context.Context, accountID string) error {
	rt, err := sm.GetRuntime(ctx, accountID)
	if err != nil {
		return err
	}
	rt.ConsecutiveFails++
	rt.LastUsedAt = time.Now()
	total := float64(100)
	rt.SuccessRate = (rt.SuccessRate * (total - 1)) / total
	return sm.SetRuntime(ctx, rt)
}

func (sm *StateManager) SetStatus(ctx context.Context, accountID, status string) error {
	return sm.rdb.HSet(ctx, sm.key(accountID), "status", status).Err()
}
