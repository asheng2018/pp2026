package paypage

import (
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/gateway/middleware"
	"github.com/ab-payment-system/internal/gateway/token"
)

type Handler struct {
	tokenGen    *token.Generator
	tmpl        *template.Template
}

type PageData struct {
	PayToken       string
	Gateway        string
	Amount         string
	OrderID        string
	BSiteDomain    string
	ClientID       string
	PublishableKey string
}

func NewHandler(tokenGen *token.Generator) *Handler {
	tmpl := template.Must(template.New("pay").Parse(payPageTemplate))
	return &Handler{
		tokenGen: tokenGen,
		tmpl:     tmpl,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	payToken, err := h.tokenGen.Validate(tokenStr)
	if err != nil {
		log.Warn().Err(err).Msg("invalid payment token")
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	middleware.SetPaymentCSP(w, payToken.Gateway)
	middleware.SetNoIndex(w)
	middleware.SetSecurityHeaders(w)

	data := &PageData{
		PayToken:    tokenStr,
		Gateway:     payToken.Gateway,
		Amount:      payToken.Amount,
		OrderID:     payToken.OrderID,
		BSiteDomain: r.Host,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("failed to render payment page")
	}
}

const payPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex,nofollow">
<meta name="referrer" content="no-referrer">
<title>Checkout</title>
<style>
  body{margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#f5f5f5}
  #payment-container{background:#fff;padding:2rem;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1);width:100%;max-width:480px}
  #loading{text-align:center;padding:2rem;color:#666}
  #error{color:#ef4444;text-align:center;display:none;padding:1rem}
  .spinner{border:3px solid #e5e7eb;border-top-color:#3b82f6;border-radius:50%;width:30px;height:30px;animation:spin 1s linear infinite;margin:1rem auto}
  @keyframes spin{to{transform:rotate(360deg)}}
  #paypal-button-container{margin-top:1rem}
</style>
</head>
<body>
<div id="payment-container">
  <div id="loading">
    <div class="spinner"></div>
    <p>Loading secure payment form...</p>
  </div>
  <div id="payment-element"></div>
  <div id="error"></div>
</div>
<script>
  var AB_PAYMENT = {
    token: "{{.PayToken}}",
    gateway: "{{.Gateway}}",
    amount: "{{.Amount}}",
    orderId: "{{.OrderID}}",
    bSiteDomain: window.location.host
  };

  // Notify parent A-site that iframe is ready
  window.parent.postMessage({
    type: "IFRAME_READY",
    data: { token: AB_PAYMENT.token }
  }, "*");

  function notifyParent(type, data) {
    window.parent.postMessage({ type: type, data: data || {} }, "*");
  }

  function showError(msg) {
    document.getElementById("loading").style.display = "none";
    var el = document.getElementById("error");
    el.textContent = msg;
    el.style.display = "block";
  }

  function showLoading(show) {
    document.getElementById("loading").style.display = show ? "block" : "none";
  }

  // Payment gateway initialization happens here
  // The SDK loader injects PayPal/Stripe JS dynamically
  // to ensure all requests originate from B-site domain
</script>
</body>
</html>`
