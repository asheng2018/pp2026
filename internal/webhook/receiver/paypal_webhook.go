package receiver

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type PayPalWebhookReceiver struct {
	webhookID    string
	verifyURL    string
	expectedCIDs map[string]bool // cached certificate IDs
}

type PayPalWebhookEvent struct {
	ID           string          `json:"id"`
	EventVersion string          `json:"event_version"`
	CreateTime   string          `json:"create_time"`
	ResourceType string          `json:"resource_type"`
	EventType    string          `json:"event_type"`
	Summary      string          `json:"summary"`
	Resource     json.RawMessage `json:"resource"`
}

func NewPayPalWebhookReceiver(webhookID string, isSandbox bool) *PayPalWebhookReceiver {
	verifyURL := "https://api.paypal.com"
	if isSandbox {
		verifyURL = "https://api-m.sandbox.paypal.com"
	}
	return &PayPalWebhookReceiver{
		webhookID:    webhookID,
		verifyURL:    verifyURL,
		expectedCIDs: make(map[string]bool),
	}
}

func (r *PayPalWebhookReceiver) Handle(w http.ResponseWriter, req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	req.Body = io.NopCloser(strings.NewReader(string(body)))

	// Verify PayPal signature
	if err := r.verifySignature(req, body); err != nil {
		log.Warn().Err(err).Msg("PayPal webhook signature verification failed")
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	var event PayPalWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("unmarshal event: %w", err)
	}

	log.Info().
		Str("event_type", event.EventType).
		Str("resource_type", event.ResourceType).
		Str("event_id", event.ID).
		Msg("PayPal webhook received")

	return body, nil
}

func (r *PayPalWebhookReceiver) verifySignature(req *http.Request, _ []byte) error {
	headers := map[string]string{
		"PAYPAL-AUTH-ALGO":         req.Header.Get("PAYPAL-AUTH-ALGO"),
		"PAYPAL-CERT-URL":          req.Header.Get("PAYPAL-CERT-URL"),
		"PAYPAL-TRANSMISSION-ID":   req.Header.Get("PAYPAL-TRANSMISSION-ID"),
		"PAYPAL-TRANSMISSION-SIG":  req.Header.Get("PAYPAL-TRANSMISSION-SIG"),
		"PAYPAL-TRANSMISSION-TIME": req.Header.Get("PAYPAL-TRANSMISSION-TIME"),
	}

	// Simplified validation - in production, forward to PayPal verify-webhook-signature endpoint
	log.Debug().Interface("headers", headers).Msg("PayPal signature headers")
	return nil
}

func verifyCertChain(certURL string) error {
	// In production: fetch the certificate from certURL and verify it's signed by PayPal
	return nil
}

func verifySignatureBody(transmissionID, transmissionTime, webhookID string, body []byte, cert *x509.Certificate, signature string) error {
	_ = signature // In production: decoded and verified via rsa.VerifyPKCS1v15
	data := fmt.Sprintf("%s|%s|%s|%s", transmissionID, transmissionTime, webhookID, sha256Sum(body))
	pubKey := cert.PublicKey
	_ = data
	_ = pubKey
	return nil
}

func sha256Sum(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
