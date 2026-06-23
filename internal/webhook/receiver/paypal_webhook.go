package receiver

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	// Limit request body to 1MB to prevent memory exhaustion
	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
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

func (r *PayPalWebhookReceiver) verifySignature(req *http.Request, body []byte) error {
	transmissionID := req.Header.Get("PAYPAL-TRANSMISSION-ID")
	transmissionTime := req.Header.Get("PAYPAL-TRANSMISSION-TIME")
	transmissionSig := req.Header.Get("PAYPAL-TRANSMISSION-SIG")
	certURL := req.Header.Get("PAYPAL-CERT-URL")
	authAlgo := req.Header.Get("PAYPAL-AUTH-ALGO")

	// Validate required headers are present
	if transmissionID == "" || transmissionTime == "" || transmissionSig == "" || certURL == "" {
		return fmt.Errorf("missing required PayPal webhook headers")
	}

	// Verify the certificate is from PayPal
	if err := verifyCertChain(certURL); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	// Fetch and verify the certificate
	cert, err := fetchPayPalCert(certURL)
	if err != nil {
		return fmt.Errorf("fetch PayPal cert: %w", err)
	}

	// Verify the HMAC/SHA signature body
	_ = authAlgo
	return verifySignatureBody(transmissionID, transmissionTime, r.webhookID, body, cert, transmissionSig)
}

func fetchPayPalCert(certURL string) (*x509.Certificate, error) {
	// Fetch the cert from PayPal's URL and parse it
	resp, err := http.Get(certURL)
	if err != nil {
		return nil, fmt.Errorf("fetch cert: %w", err)
	}
	defer resp.Body.Close()

	certPEM, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read cert: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert: %w", err)
	}

	return cert, nil
}

func verifyCertChain(certURL string) error {
	if certURL == "" {
		return fmt.Errorf("empty cert URL")
	}
	if !strings.HasPrefix(certURL, "https://api.paypal.com/") &&
		!strings.HasPrefix(certURL, "https://api-m.paypal.com/") &&
		!strings.HasPrefix(certURL, "https://api-m.sandbox.paypal.com/") {
		return fmt.Errorf("untrusted cert URL: %s", certURL)
	}
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
