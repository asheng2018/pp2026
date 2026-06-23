package activity

import (
	"context"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type TransactionSimulator struct {
	minAmount float64
	maxAmount float64
	minDaily  int
	maxDaily  int
}

func NewTransactionSimulator(minAmount, maxAmount float64, minDaily, maxDaily int) *TransactionSimulator {
	return &TransactionSimulator{
		minAmount: minAmount,
		maxAmount: maxAmount,
		minDaily:  minDaily,
		maxDaily:  maxDaily,
	}
}

func (s *TransactionSimulator) SimulateDaily(ctx context.Context, accountID string, processPayment func(ctx context.Context, amount decimal.Decimal) error) error {
	today := time.Now().Weekday()
	// Simulate natural transaction patterns
	isWeekend := today == time.Saturday || today == time.Sunday
	numTx := s.minDaily + rand.Intn(s.maxDaily-s.minDaily+1)
	if isWeekend {
		numTx = numTx * 2 / 3 // fewer on weekends
	}

	log.Info().Str("account", accountID).Int("transactions", numTx).Msg("simulating daily transactions")

	for i := 0; i < numTx; i++ {
		amount := decimal.NewFromFloat(s.minAmount + rand.Float64()*(s.maxAmount-s.minAmount))
		// Vary amounts to look natural
		if rand.Float64() > 0.7 {
			amount = decimal.NewFromFloat(s.maxAmount * 1.5) // occasional larger purchase
		}
		if err := processPayment(ctx, amount); err != nil {
			log.Warn().Err(err).Int("tx", i).Msg("simulated transaction failed")
			continue
		}
		// Random delay between transactions
		delay := time.Duration(5+rand.Intn(30)) * time.Minute
		time.Sleep(delay)
	}

	return nil
}
