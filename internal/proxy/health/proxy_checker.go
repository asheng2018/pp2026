package health

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ab-payment-system/internal/proxy/model"
)

type ProxyChecker struct {
	timeout       time.Duration
	paypalTestURL string
	stripeTestURL string
}

func NewProxyChecker(timeout time.Duration) *ProxyChecker {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &ProxyChecker{
		timeout:       timeout,
		paypalTestURL: "https://api-m.paypal.com",
		stripeTestURL: "https://api.stripe.com",
	}
}

func (pc *ProxyChecker) Check(p *model.Proxy) *model.ProxyHealth {
	h := &model.ProxyHealth{}

	proxyURL, err := url.Parse(fmt.Sprintf("%s://%s:%d", p.Protocol, p.Host, p.Port))
	if err != nil {
		return h
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: pc.timeout,
	}

	// Latency test via httpbin
	start := time.Now()
	resp, err := client.Get("https://httpbin.org/ip")
	if err != nil {
		return h
	}
	h.Latency = int(time.Since(start).Milliseconds())
	resp.Body.Close()

	// PayPal accessibility test
	if resp, err := client.Get(pc.paypalTestURL); err == nil && resp.StatusCode < 500 {
		h.PayPalAccessible = true
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	// Stripe accessibility test
	if resp, err := client.Get(pc.stripeTestURL); err == nil && resp.StatusCode < 500 {
		h.StripeAccessible = true
		if resp.Body != nil {
			resp.Body.Close()
		}
	}

	h.IsHealthy = h.Latency > 0 && h.Latency < 10000
	return h
}
