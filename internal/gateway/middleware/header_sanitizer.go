package middleware

import (
	"net/http"
	"strings"
)

var AllowedHeaders = []string{
	"Content-Type", "Accept", "Accept-Language",
	"User-Agent", "Cache-Control",
}

var BlockedHeaders = []string{
	"Referer", "Origin", "X-Forwarded-For", "Via",
	"X-Real-Ip", "Cf-Connecting-Ip", "True-Client-Ip",
}

func SanitizeHeaders(headers http.Header) http.Header {
	clean := make(http.Header)
	for _, k := range AllowedHeaders {
		if v := headers.Get(k); v != "" {
			clean.Set(k, v)
		}
	}
	return clean
}

func CleanReferrer(headers http.Header) http.Header {
	headers.Del("Referer")
	headers.Set("Referrer-Policy", "no-referrer")
	return headers
}

func HasBlockedHeaders(headers http.Header) bool {
	for _, k := range BlockedHeaders {
		if headers.Get(k) != "" {
			return true
		}
	}
	return false
}

func StripTrackingParams(urlStr string) string {
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign",
		"utm_content", "utm_term", "fbclid", "gclid", "ref",
	}
	for _, p := range trackingParams {
		prefix := p + "="
		if idx := strings.Index(urlStr, prefix); idx >= 0 {
			end := strings.Index(urlStr[idx:], "&")
			if end < 0 {
				if idx > 0 {
					urlStr = urlStr[:idx-1]
				} else {
					urlStr = urlStr[:idx]
				}
				break
			}
			if idx > 0 {
				urlStr = urlStr[:idx] + urlStr[idx+end+1:]
			} else {
				urlStr = urlStr[idx+end+1:]
			}
		}
	}
	return urlStr
}
