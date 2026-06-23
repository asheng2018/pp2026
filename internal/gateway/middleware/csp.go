package middleware

import "net/http"

func SetPaymentCSP(w http.ResponseWriter, gateway string) {
	var csp string
	switch gateway {
	case "paypal":
		csp = "default-src 'self'; " +
			"script-src 'self' https://www.paypal.com https://www.paypalobjects.com; " +
			"frame-src https://www.paypal.com; " +
			"connect-src https://api.paypal.com; " +
			"img-src 'self' https://www.paypalobjects.com; " +
			"style-src 'self' 'unsafe-inline'"
	case "stripe":
		csp = "default-src 'self'; " +
			"script-src 'self' https://js.stripe.com; " +
			"frame-src https://js.stripe.com https://hooks.stripe.com; " +
			"connect-src https://api.stripe.com; " +
			"img-src 'self' https://*.stripe.com; " +
			"style-src 'self' 'unsafe-inline'"
	default:
		csp = "default-src 'self'"
	}
	w.Header().Set("Content-Security-Policy", csp)
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func SetNoIndex(w http.ResponseWriter) {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
}

func SetSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
}
