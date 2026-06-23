package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type ExchangeRate struct {
	BaseCurrency   string          `json:"base_currency"`
	TargetCurrency string          `json:"target_currency"`
	Rate           decimal.Decimal `json:"rate"`
	Source         string          `json:"source"`
	FetchedAt      time.Time       `json:"fetched_at"`
}

type RateProvider interface {
	Name() string
	FetchRates(ctx context.Context, baseCurrency string) (map[string]decimal.Decimal, error)
}

type OpenExchangeRatesProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewOpenExchangeRatesProvider(apiKey string) *OpenExchangeRatesProvider {
	return &OpenExchangeRatesProvider{
		apiKey:   apiKey,
		endpoint: "https://openexchangerates.org/api",
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OpenExchangeRatesProvider) Name() string { return "openexchangerates" }

func (p *OpenExchangeRatesProvider) FetchRates(ctx context.Context, baseCurrency string) (map[string]decimal.Decimal, error) {
	url := fmt.Sprintf("%s/latest.json?app_id=%s&base=%s", p.endpoint, p.apiKey, baseCurrency)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch rates: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode rates: %w", err)
	}

	rates := make(map[string]decimal.Decimal)
	for currency, rate := range result.Rates {
		rates[currency] = decimal.NewFromFloat(rate)
	}
	return rates, nil
}
