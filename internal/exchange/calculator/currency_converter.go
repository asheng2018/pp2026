package calculator

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/exchange/provider"
)

type CurrencyConverter struct {
	mu        sync.RWMutex
	rates     map[string]map[string]decimal.Decimal // base -> {target: rate}
	providers []provider.RateProvider
	markup    decimal.Decimal // additional markup percentage
}

func NewCurrencyConverter(providers []provider.RateProvider, markup decimal.Decimal) *CurrencyConverter {
	if markup.IsZero() {
		markup = decimal.NewFromFloat(0.02) // default 2% markup
	}
	return &CurrencyConverter{
		rates:     make(map[string]map[string]decimal.Decimal),
		providers: providers,
		markup:    markup,
	}
}

func (c *CurrencyConverter) Convert(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}

	rate, err := c.GetRate(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}

	// Apply markup
	effectiveRate := rate.Mul(decimal.NewFromInt(1).Add(c.markup))
	converted := amount.Mul(effectiveRate)

	return converted.Round(2), nil
}

func (c *CurrencyConverter) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	c.mu.RLock()
	if baseRates, ok := c.rates[from]; ok {
		if rate, ok := baseRates[to]; ok {
			c.mu.RUnlock()
			return rate, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from providers
	return c.fetchRate(ctx, from, to)
}

func (c *CurrencyConverter) fetchRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	for _, prov := range c.providers {
		rates, err := prov.FetchRates(ctx, from)
		if err != nil {
			log.Warn().Str("provider", prov.Name()).Err(err).Msg("failed to fetch rates")
			continue
		}

		c.mu.Lock()
		c.rates[from] = rates
		c.mu.Unlock()

		if rate, ok := rates[to]; ok {
			return rate, nil
		}
	}
	return decimal.Zero, fmt.Errorf("rate not found: %s->%s", from, to)
}

func (c *CurrencyConverter) RefreshRates(ctx context.Context) error {
	// Refresh all known currency pairs
	for _, base := range []string{"USD", "EUR", "GBP"} {
		_, err := c.fetchRate(ctx, base, "")
		if err != nil {
			log.Warn().Str("base", base).Err(err).Msg("rate refresh failed")
		}
	}
	return nil
}
