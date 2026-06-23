package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AdminClaims carries the JWT payload for admin authentication.
type AdminClaims struct {
	jwt.RegisteredClaims
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// TokenService creates and validates admin JWT tokens.
type TokenService struct {
	secret []byte
	expiry time.Duration
}

// NewTokenService creates a new TokenService with the given secret and expiry.
func NewTokenService(secret string, expiry time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Generate creates a signed JWT for the given admin user.
func (s *TokenService) Generate(adminID, username, role string) (string, *AdminClaims, error) {
	now := time.Now().UTC()
	claims := &AdminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   adminID,
			Issuer:    "ab-payment-system",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
			ID:        uuid.New().String(),
		},
		AdminID:  adminID,
		Username: username,
		Role:     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", nil, err
	}
	return tokenString, claims, nil
}

// Validate parses and validates a JWT string. Returns the claims on success,
// or an error if the token is invalid, expired, or malformed.
func (s *TokenService) Validate(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return s.secret, nil
		},
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
