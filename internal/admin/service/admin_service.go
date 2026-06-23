package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ab-payment-system/internal/admin/model"
	"github.com/ab-payment-system/internal/admin/repository"
)

// Common errors returned by AdminService.
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenBlacklisted   = errors.New("token has been revoked")
)

// LoginResult is returned on successful authentication.
type LoginResult struct {
	Token     string    `json:"token"`
	AdminID   string    `json:"admin_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AdminService provides admin authentication business logic.
type AdminService struct {
	repo      *repository.AdminRepository
	tokenSvc  *TokenService
	blacklist *TokenBlacklist
}

// NewAdminService creates a new AdminService.
func NewAdminService(repo *repository.AdminRepository, tokenSvc *TokenService, blacklist *TokenBlacklist) *AdminService {
	return &AdminService{
		repo:      repo,
		tokenSvc:  tokenSvc,
		blacklist: blacklist,
	}
}

// Login verifies credentials and returns a JWT token pair.
func (s *AdminService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login time (best-effort; log failure but don't block login).
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Logging is done by the caller/handler.
		// We don't fail the login for this.
	}

	tokenStr, claims, err := s.tokenSvc.Generate(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token:     tokenStr,
		AdminID:   user.ID,
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// Logout revokes the given token by adding its jti to the blacklist.
func (s *AdminService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.tokenSvc.Validate(tokenString)
	if err != nil {
		return err
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token already expired; nothing to blacklist.
		return nil
	}

	return s.blacklist.Blacklist(ctx, claims.ID, ttl)
}

// GetCurrentUser validates the token and returns the current admin user.
// It fetches from the database to ensure fresh data (role changes, etc.).
func (s *AdminService) GetCurrentUser(ctx context.Context, tokenString string) (*model.AdminUser, error) {
	claims, err := s.tokenSvc.Validate(tokenString)
	if err != nil {
		return nil, err
	}

	// Check blacklist
	if blacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID); err != nil {
		return nil, err
	} else if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	user, err := s.repo.FindByUsername(ctx, claims.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// ValidateToken is used by the middleware: validates the JWT and checks the blacklist.
// Returns the claims on success.
func (s *AdminService) ValidateToken(ctx context.Context, tokenString string) (*AdminClaims, error) {
	claims, err := s.tokenSvc.Validate(tokenString)
	if err != nil {
		return nil, err
	}

	if blacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID); err != nil {
		return nil, err
	} else if blacklisted {
		return nil, ErrTokenBlacklisted
	}

	return claims, nil
}
