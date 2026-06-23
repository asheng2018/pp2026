package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type FarmingPlan struct {
	AccountID         string
	Phase             string // registration | warming | active | retirement
	StartDate         time.Time
	EndDate           time.Time
	DailyTransactions int
	WeeklyBudget      float64
	Activities        []FarmingActivity
}

type FarmingActivity struct {
	Type     string // transaction | social | review | update
	Schedule string // cron expression
	Config   map[string]interface{}
}

type FarmingScheduler struct {
	cron  *cron.Cron
	plans map[string]*FarmingPlan
}

func NewFarmingScheduler() *FarmingScheduler {
	return &FarmingScheduler{
		cron:  cron.New(),
		plans: make(map[string]*FarmingPlan),
	}
}

func (fs *FarmingScheduler) AddPlan(plan *FarmingPlan) {
	fs.plans[plan.AccountID] = plan
	log.Info().Str("account", plan.AccountID).Str("phase", plan.Phase).Msg("farming plan added")
}

func (fs *FarmingScheduler) RemovePlan(accountID string) {
	delete(fs.plans, accountID)
}

func (fs *FarmingScheduler) Start() error {
	// Run daily farming activities at random times during business hours
	_, err := fs.cron.AddFunc("0 10 * * *", func() {
		ctx := context.Background()
		log.Info().Msg("running scheduled farming activities")
		for _, plan := range fs.plans {
			fs.executePlan(ctx, plan)
		}
	})
	if err != nil {
		return err
	}

	fs.cron.Start()
	log.Info().Msg("farming scheduler started")
	return nil
}

func (fs *FarmingScheduler) executePlan(ctx context.Context, plan *FarmingPlan) {
	log.Info().
		Str("account", plan.AccountID).
		Str("phase", plan.Phase).
		Msg("executing farming plan")
	// Execute activities based on the plan phase
}

func (fs *FarmingScheduler) Stop() {
	fs.cron.Stop()
}
