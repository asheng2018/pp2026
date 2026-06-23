package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code, message string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode}
}

func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode, Err: err}
}

// Common errors
var (
	ErrNoAccountAvailable = New("NO_ACCOUNT", "no payment account available", 503)
	ErrAccountNotFound    = New("ACCOUNT_NOT_FOUND", "account not found", 404)
	ErrOrderNotFound      = New("ORDER_NOT_FOUND", "order not found", 404)
	ErrMerchantNotFound   = New("MERCHANT_NOT_FOUND", "merchant not found", 404)
	ErrInvalidToken       = New("INVALID_TOKEN", "invalid payment token", 401)
	ErrLimitExceeded      = New("LIMIT_EXCEEDED", "account limit exceeded", 429)
	ErrCircuitOpen        = New("CIRCUIT_OPEN", "circuit breaker is open", 503)
	ErrProxyUnavailable   = New("PROXY_UNAVAILABLE", "proxy is not available", 503)
)
