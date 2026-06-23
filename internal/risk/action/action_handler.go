package action

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/risk/model"
)

// Result represents the outcome of processing a risk action.
type Result struct {
	Allowed   bool   `json:"allowed"`
	Throttled bool   `json:"throttled"`
	Reason    string `json:"reason"`
	NewAccountID string `json:"new_account_id,omitempty"`
}

type ActionHandler struct {
	switchAccountFn func(ctx context.Context, merchantID, currentAccountID string) (string, error)
}

func NewActionHandler(
	switchAccountFn func(ctx context.Context, merchantID, currentAccountID string) (string, error),
) *ActionHandler {
	return &ActionHandler{
		switchAccountFn: switchAccountFn,
	}
}

func (h *ActionHandler) Execute(ctx context.Context, event *model.RiskEvent, merchantID string) (*Result, error) {
	switch event.Action {
	case model.ActionAllow:
		return &Result{Allowed: true}, nil

	case model.ActionBlock:
		log.Ctx(ctx).Warn().
			Str("rule", event.RuleName).
			Str("reason", event.Reason).
			Msg("order blocked by risk engine")
		return &Result{Allowed: false, Reason: event.Reason}, nil

	case model.ActionFlag:
		log.Ctx(ctx).Warn().
			Str("rule", event.RuleName).
			Msg("order flagged for manual review")
		return &Result{Allowed: true, Reason: "flagged for review"}, nil

	case model.ActionThrottle:
		log.Ctx(ctx).Warn().
			Str("rule", event.RuleName).
			Msg("throttling applied")
		return &Result{Allowed: true, Throttled: true, Reason: event.Reason}, nil

	case model.ActionSwitchAccount:
		log.Ctx(ctx).Warn().
			Str("rule", event.RuleName).
			Str("account", event.AccountID).
			Msg("switching account due to risk")
		newAccountID, err := h.switchAccountFn(ctx, merchantID, event.AccountID)
		if err != nil {
			return &Result{Allowed: false, Reason: "account switch failed"}, err
		}
		return &Result{Allowed: true, NewAccountID: newAccountID}, nil

	case model.ActionNotify:
		log.Ctx(ctx).Warn().
			Str("rule", event.RuleName).
			Msg("risk notification triggered")
		return &Result{Allowed: true}, nil

	default:
		return &Result{Allowed: true}, nil
	}
}
