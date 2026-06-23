package event

import (
	"context"
	"encoding/json"
	"time"
)

type PaymentEvent struct {
	ID        string          `json:"id"`
	Gateway   string          `json:"gateway"`
	EventType string          `json:"event_type"`
	OrderID   string          `json:"order_id,omitempty"`
	Amount    string          `json:"amount,omitempty"`
	Currency  string          `json:"currency,omitempty"`
	Status    string          `json:"status,omitempty"`
	RawData   json.RawMessage `json:"raw_data,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type EventHandler interface {
	Handle(ctx context.Context, event *PaymentEvent) error
}

// PayPal event types
const (
	PayPalCheckoutOrderApproved   = "CHECKOUT.ORDER.APPROVED"
	PayPalCheckoutOrderCompleted  = "CHECKOUT.ORDER.COMPLETED"
	PayPalPaymentCaptureCompleted = "PAYMENT.CAPTURE.COMPLETED"
	PayPalPaymentCaptureDenied    = "PAYMENT.CAPTURE.DENIED"
	PayPalPaymentCaptureRefunded  = "PAYMENT.CAPTURE.REFUNDED"
	PayPalDisputeCreated          = "CUSTOMER.DISPUTE.CREATED"
	PayPalDisputeResolved         = "CUSTOMER.DISPUTE.RESOLVED"
)

// Stripe event types
const (
	StripePaymentIntentSucceeded = "payment_intent.succeeded"
	StripePaymentIntentFailed    = "payment_intent.payment_failed"
	StripeChargeRefunded         = "charge.refunded"
	StripeDisputeCreated         = "charge.dispute.created"
	StripeDisputeClosed          = "charge.dispute.closed"
)
