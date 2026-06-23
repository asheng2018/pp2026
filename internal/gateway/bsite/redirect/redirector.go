package redirect

import (
	"html/template"
	"net/http"
)

type ResultData struct {
	Status        string
	OrderID       string
	Amount        string
	Message       string
	RedirectURL   string
	RedirectDelay int
}

var resultTmpl = template.Must(template.New("result").Parse(resultTemplate))

func RenderResult(w http.ResponseWriter, status, orderID, amount, message, redirectURL string) {
	delay := 0
	if redirectURL != "" {
		delay = 3
	}
	data := ResultData{
		Status:        status,
		OrderID:       orderID,
		Amount:        amount,
		Message:       message,
		RedirectURL:   redirectURL,
		RedirectDelay: delay,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	resultTmpl.Execute(w, data)
}

func HandleSuccess(w http.ResponseWriter, r *http.Request, orderID, amount, redirectURL string) {
	msg := "Payment completed successfully."
	if redirectURL != "" {
		msg += " You will be redirected shortly."
	}
	RenderResult(w, "success", orderID, amount, msg, redirectURL)
}

func HandleFailure(w http.ResponseWriter, r *http.Request, orderID, reason, redirectURL string) {
	msg := "Payment failed."
	if reason != "" {
		msg = "Payment failed: " + reason
	}
	RenderResult(w, "failed", orderID, "", msg, redirectURL)
}

const resultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="robots" content="noindex,nofollow">
<meta name="referrer" content="no-referrer">
{{if .RedirectURL}}<meta http-equiv="refresh" content="{{.RedirectDelay}};url={{.RedirectURL}}">{{end}}
<title>Payment {{.Status}}</title>
<style>
  body{font-family:-apple-system,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f5f5f5}
  .result{text-align:center;padding:2rem;background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1)}
  .success{color:#10b981} .failed{color:#ef4444}
  .spinner{border:3px solid #e5e7eb;border-top-color:#3b82f6;border-radius:50%;width:30px;height:30px;animation:spin 1s linear infinite;margin:1rem auto}
  @keyframes spin{to{transform:rotate(360deg)}}
</style>
</head>
<body>
<div class="result">
  {{if eq .Status "success"}}
    <h1 class="success">&#10003; Payment Successful</h1>
  {{else}}
    <h1 class="failed">&#10007; Payment Failed</h1>
  {{end}}
  <p>{{.Message}}</p>
  {{if .RedirectURL}}
    <div class="spinner"></div>
    <p>Redirecting in {{.RedirectDelay}} seconds...</p>
  {{end}}
</div>
<script>
  if (window.parent && window.parent !== window) {
    window.parent.postMessage({
      type: "{{if eq .Status "success"}}PAYMENT_COMPLETED{{else}}PAYMENT_FAILED{{end}}",
      data: { orderId: "{{.OrderID}}", amount: "{{.Amount}}" }
    }, "*");
  }
</script>
</body>
</html>`
