package schedule

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/settlement/engine"
)

type SettlementJob struct {
	engine *engine.SettlementEngine
	cron   *cron.Cron
}

func NewSettlementJob(settlementEngine *engine.SettlementEngine) *SettlementJob {
	return &SettlementJob{
		engine: settlementEngine,
		cron:   cron.New(),
	}
}

func (j *SettlementJob) Start() error {
	// Run settlement processing every hour
	_, err := j.cron.AddFunc("0 * * * *", func() {
		ctx := context.Background()
		log.Info().Msg("running settlement processing job")
		processed, failed, err := j.engine.ProcessPending(ctx)
		if err != nil {
			log.Error().Err(err).Msg("settlement processing failed")
			return
		}
		log.Info().
			Int("processed", processed).
			Int("failed", failed).
			Msg("settlement processing completed")
	})
	if err != nil {
		return err
	}

	// Generate daily settlements at midnight
	_, err = j.cron.AddFunc("0 0 * * *", func() {
		_ = context.Background()
		now := time.Now()
		cycleStart := now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		cycleEnd := cycleStart.Add(24 * time.Hour)

		log.Info().
			Time("cycle_start", cycleStart).
			Time("cycle_end", cycleEnd).
			Msg("generating daily settlements")
		// TODO: Iterate all merchants and generate settlements
	})

	j.cron.Start()
	log.Info().Msg("settlement scheduler started")
	return nil
}

func (j *SettlementJob) Stop() {
	j.cron.Stop()
	log.Info().Msg("settlement scheduler stopped")
}
