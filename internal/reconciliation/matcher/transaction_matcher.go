package matcher

import (
	"context"
	"math"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type GatewayTransaction struct {
	TxID        string
	OrderID     string
	Amount      decimal.Decimal
	Fee         decimal.Decimal
	Currency    string
	Status      string
	ProcessedAt time.Time
}

type SystemTransaction struct {
	OrderID string
	Amount  decimal.Decimal
	Gateway string
	Status  string
	PaidAt  *time.Time
}

type MatchResult struct {
	SystemTx   *SystemTransaction
	GatewayTx  *GatewayTransaction
	Matched    bool
	DiffAmount decimal.Decimal
	Reason     string
}

type TransactionMatcher struct {
	amountTolerance float64
	timeWindow      time.Duration
}

func NewTransactionMatcher(amountTolerance float64, timeWindow time.Duration) *TransactionMatcher {
	if amountTolerance <= 0 {
		amountTolerance = 0.01 // $0.01 tolerance
	}
	if timeWindow <= 0 {
		timeWindow = 24 * time.Hour
	}
	return &TransactionMatcher{
		amountTolerance: amountTolerance,
		timeWindow:      timeWindow,
	}
}

func (m *TransactionMatcher) Match(ctx context.Context, systemTxs []*SystemTransaction, gatewayTxs []*GatewayTransaction) []*MatchResult {
	var results []*MatchResult

	// Index gateway transactions by order ID for exact match
	gwIndex := make(map[string]*GatewayTransaction)
	for _, gw := range gatewayTxs {
		if gw.OrderID != "" {
			gwIndex[gw.OrderID] = gw
		}
	}

	matchedGW := make(map[string]bool)

	for _, sys := range systemTxs {
		result := &MatchResult{SystemTx: sys}

		// Try exact match by order ID
		if gw, ok := gwIndex[sys.OrderID]; ok {
			result.GatewayTx = gw
			diff := sys.Amount.Sub(gw.Amount).Abs()
			result.DiffAmount = diff

			if diff.InexactFloat64() <= m.amountTolerance {
				result.Matched = true
				result.Reason = "exact_match"
			} else {
				result.Matched = false
				result.Reason = "amount_mismatch"
			}
			matchedGW[gw.TxID] = true
			results = append(results, result)
			continue
		}

		// Try fuzzy match by amount and time window
		bestMatch := m.fuzzyMatch(sys, gatewayTxs, matchedGW)
		if bestMatch != nil {
			result.GatewayTx = bestMatch
			result.Matched = true
			result.Reason = "fuzzy_match"
			matchedGW[bestMatch.TxID] = true
		} else {
			result.Matched = false
			result.Reason = "no_gateway_match"
		}

		results = append(results, result)
	}

	// Add unmatched gateway transactions
	for _, gw := range gatewayTxs {
		if !matchedGW[gw.TxID] {
			results = append(results, &MatchResult{
				GatewayTx: gw,
				Matched:   false,
				Reason:    "no_system_match",
			})
		}
	}

	log.Info().Int("total", len(results)).Msg("reconciliation matching completed")
	return results
}

func (m *TransactionMatcher) fuzzyMatch(sys *SystemTransaction, gatewayTxs []*GatewayTransaction, matched map[string]bool) *GatewayTransaction {
	var best *GatewayTransaction
	var bestDiff float64 = math.MaxFloat64

	for _, gw := range gatewayTxs {
		if matched[gw.TxID] {
			continue
		}
		diff := sys.Amount.Sub(gw.Amount).Abs().InexactFloat64()
		if diff <= m.amountTolerance && diff < bestDiff {
			// Check time proximity if sys has PaidAt
			if sys.PaidAt != nil {
				timeDiff := gw.ProcessedAt.Sub(*sys.PaidAt).Abs()
				if timeDiff <= m.timeWindow {
					best = gw
					bestDiff = diff
				}
			} else {
				best = gw
				bestDiff = diff
			}
		}
	}
	return best
}
