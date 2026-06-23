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

func StripTrackingParams(rawURL string) string {
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign",
		"utm_content", "utm_term", "fbclid", "gclid", "ref",
	}
	for _, p := range trackingParams {
		rawURL = stripParam(rawURL, p)
	}
	return rawURL
}

// stripParam removes a single query parameter and its value from a URL.
func stripParam(rawURL, param string) string {
	prefix := param + "="
	for {
		idx := strings.Index(rawURL, prefix)
		if idx < 0 {
			return rawURL
		}
		// Find the start of this parameter (after ? or &)
		paramStart := idx
		if idx > 0 && rawURL[idx-1] == '?' {
			paramStart = idx
		} else if idx > 1 && rawURL[idx-2:idx] == "? " {
			paramStart = idx
		}
		// Find the end of this parameter's value
		end := idx + len(prefix)
		for end < len(rawURL) && rawURL[end] != '&' {
			end++
		}
		// Remove from rawURL: cut from paramStart to end (inclusive of trailing &)
		if end < len(rawURL) {
			rawURL = rawURL[:paramStart] + rawURL[end+1:]
		} else {
			rawURL = rawURL[:paramStart]
			if paramStart > 0 && rawURL[paramStart-1] == '?' {
				rawURL = rawURL[:paramStart-1]
			}
		}
	}
}
