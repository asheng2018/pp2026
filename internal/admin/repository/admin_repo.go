package repository

import (
	"context"
	"database/sql"

	"github.com/ab-payment-system/internal/admin/model"
)

// AdminRepository provides data access for admin_users and audit_logs.
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// FindByUsername returns an AdminUser by username. Returns nil, nil if not found.
func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, totp_secret, totp_enabled, last_login_at, created_at
		 FROM admin_users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.TOTPSecret, &u.TOTPEnabled, &u.LastLoginAt, &u.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateLastLogin sets the last_login_at timestamp to NOW() for the given admin.
func (r *AdminRepository) UpdateLastLogin(ctx context.Context, adminID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE admin_users SET last_login_at = NOW() WHERE id = $1`, adminID)
	return err
}

// CreateAuditLog inserts a new audit log entry.
func (r *AdminRepository) CreateAuditLog(ctx context.Context, entry *model.AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.AdminID, entry.Action, entry.ResourceType, entry.ResourceID,
		entry.OldValue, entry.NewValue, entry.IPAddress, entry.UserAgent)
	return err
}
