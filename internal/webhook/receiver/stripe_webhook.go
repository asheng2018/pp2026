package receiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeWebhookReceiver struct {
	endpointSecret string
}

type StripeEvent struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func NewStripeWebhookReceiver(endpointSecret string) *StripeWebhookReceiver {
	return &StripeWebhookReceiver{endpointSecret: endpointSecret}
}

func (r *StripeWebhookReceiver) Handle(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Verify Stripe signature
	sigHeader := req.Header.Get("Stripe-Signature")
	_, err = webhook.ConstructEvent(body, sigHeader, r.endpointSecret)
	if err != nil {
		log.Warn().Err(err).Msg("Stripe webhook signature verification failed")
		return nil, fmt.Errorf("stripe signature verification: %w", err)
	}

	var event StripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}

	log.Info().
		Str("event_type", event.Type).
		Str("event_id", event.ID).
		Msg("Stripe webhook received")

	return body, nil
}
