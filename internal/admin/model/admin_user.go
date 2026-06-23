package model

import "time"

// AdminUser represents a row in the admin_users table.
type AdminUser struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         string     `json:"role" db:"role"`
	TOTPSecret   *string    `json:"-" db:"totp_secret"`
	TOTPEnabled  bool       `json:"totp_enabled" db:"totp_enabled"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog represents a row in the audit_logs table.
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	AdminID      string    `json:"admin_id" db:"admin_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   *string   `json:"resource_id,omitempty" db:"resource_id"`
	OldValue     *string   `json:"old_value,omitempty" db:"old_value"`
	NewValue     *string   `json:"new_value,omitempty" db:"new_value"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
