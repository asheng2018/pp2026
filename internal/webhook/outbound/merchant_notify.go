package outbound

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type NotifyPayload struct {
	OrderID   string `json:"order_id"`
	OrderNo   string `json:"order_no"`
	Status    string `json:"status"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Gateway   string `json:"gateway"`
	PaidAt    string `json:"paid_at,omitempty"`
	Signature string `json:"signature"`
}

type MerchantNotifier struct {
	httpClient *http.Client
	retryMax   int
	retryDelay time.Duration
}

func NewMerchantNotifier(retryMax int, retryDelay time.Duration) *MerchantNotifier {
	return &MerchantNotifier{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		retryMax:   retryMax,
		retryDelay: retryDelay,
	}
}

func (n *MerchantNotifier) Notify(ctx context.Context, webhookURL, secret string, payload *NotifyPayload) error {
	// Sign payload
	sig := signPayload(payload, secret)
	payload.Signature = sig

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastErr error
	for i := 0; i <= n.retryMax; i++ {
		if i > 0 {
			delay := n.retryDelay * time.Duration(1<<(i-1)) // exponential backoff
			log.Debug().Int("retry", i).Dur("delay", delay).Msg("retrying webhook")
			time.Sleep(delay)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(data))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-AB-Signature", sig)

		resp, err := n.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Info().Str("order_id", payload.OrderID).Str("status", payload.Status).Msg("merchant notified")
			return nil
		}
		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return fmt.Errorf("webhook notification failed after %d retries: %w", n.retryMax, lastErr)
}

func signPayload(payload *NotifyPayload, secret string) string {
	data, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
