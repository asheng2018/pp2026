package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/google/uuid"
)

func NewID() string             { return uuid.New().String() }
func NewShortID() string        { return uuid.New().String()[:8] }
func RandomHex(n int) string    { b := make([]byte, n); rand.Read(b); return hex.EncodeToString(b) }
func HMACSHA256(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
func MaskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
func Ptr[T any](v T) *T       { return &v }
func Val[T any](p *T) T       { var zero T; if p == nil { return zero }; return *p }
