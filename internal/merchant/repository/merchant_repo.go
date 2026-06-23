package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/ab-payment-system/internal/merchant/model"
)

// MerchantRepo is the PostgreSQL implementation of the merchant repository.
type MerchantRepo struct {
	db *sql.DB
}

// NewMerchantRepo creates a new MerchantRepo.
func NewMerchantRepo(db *sql.DB) *MerchantRepo {
	return &MerchantRepo{db: db}
}

// FindByID retrieves a merchant by UUID.
func (r *MerchantRepo) FindByID(ctx context.Context, id string) (*model.Merchant, error) {
	m := &model.Merchant{}
	var riskJSON, feeJSON, settleJSON, metaJSON []byte
	var email, routing, group sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, status, risk_profile, fee_config, settlement_config,
		       routing_mode, account_group, metadata, created_at, updated_at
		FROM merchants WHERE id = $1`, id,
	).Scan(&m.ID, &m.Name, &email, &m.Status, &riskJSON, &feeJSON, &settleJSON,
		&routing, &group, &metaJSON, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if email.Valid {
		m.Email = email.String
	}
	if routing.Valid {
		m.RoutingMode = routing.String
	}
	if group.Valid {
		m.AccountGroup = group.String
	}
	if riskJSON != nil {
		json.Unmarshal(riskJSON, &m.RiskProfile)
	}
	if feeJSON != nil {
		json.Unmarshal(feeJSON, &m.FeeConfig)
	}
	if settleJSON != nil {
		json.Unmarshal(settleJSON, &m.Settlement)
	}
	if metaJSON != nil {
		json.Unmarshal(metaJSON, &m.Metadata)
	}
	return m, nil
}

// FindByEmail retrieves a merchant by email.
func (r *MerchantRepo) FindByEmail(ctx context.Context, email string) (*model.Merchant, error) {
	m := &model.Merchant{}
	var riskJSON, feeJSON, settleJSON, metaJSON []byte
	var emailDB, routing, group sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, status, risk_profile, fee_config, settlement_config,
		       routing_mode, account_group, metadata, created_at, updated_at
		FROM merchants WHERE email = $1`, email,
	).Scan(&m.ID, &m.Name, &emailDB, &m.Status, &riskJSON, &feeJSON, &settleJSON,
		&routing, &group, &metaJSON, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if emailDB.Valid {
		m.Email = emailDB.String
	}
	if routing.Valid {
		m.RoutingMode = routing.String
	}
	if group.Valid {
		m.AccountGroup = group.String
	}
	if riskJSON != nil {
		json.Unmarshal(riskJSON, &m.RiskProfile)
	}
	if feeJSON != nil {
		json.Unmarshal(feeJSON, &m.FeeConfig)
	}
	if settleJSON != nil {
		json.Unmarshal(settleJSON, &m.Settlement)
	}
	if metaJSON != nil {
		json.Unmarshal(metaJSON, &m.Metadata)
	}
	return m, nil
}

// Create inserts a new merchant.
func (r *MerchantRepo) Create(ctx context.Context, m *model.Merchant) error {
	riskJSON, _ := json.Marshal(m.RiskProfile)
	feeJSON, _ := json.Marshal(m.FeeConfig)
	settleJSON, _ := json.Marshal(m.Settlement)
	metaJSON, _ := json.Marshal(m.Metadata)

	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return r.db.QueryRowContext(ctx, `
		INSERT INTO merchants (id, name, email, status, risk_profile, fee_config,
			settlement_config, routing_mode, account_group, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at, updated_at`,
		m.ID, m.Name, sql.NullString{String: m.Email, Valid: m.Email != ""},
		m.Status, riskJSON, feeJSON, settleJSON,
		sql.NullString{String: m.RoutingMode, Valid: m.RoutingMode != ""},
		sql.NullString{String: m.AccountGroup, Valid: m.AccountGroup != ""},
		metaJSON,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
}

// Update modifies an existing merchant.
func (r *MerchantRepo) Update(ctx context.Context, m *model.Merchant) error {
	riskJSON, _ := json.Marshal(m.RiskProfile)
	feeJSON, _ := json.Marshal(m.FeeConfig)
	settleJSON, _ := json.Marshal(m.Settlement)
	metaJSON, _ := json.Marshal(m.Metadata)

	_, err := r.db.ExecContext(ctx, `
		UPDATE merchants SET name=$1, email=$2, status=$3, risk_profile=$4,
			fee_config=$5, settlement_config=$6, routing_mode=$7,
			account_group=$8, metadata=$9, updated_at=NOW()
		WHERE id=$10`,
		m.Name, sql.NullString{String: m.Email, Valid: m.Email != ""},
		m.Status, riskJSON, feeJSON, settleJSON,
		sql.NullString{String: m.RoutingMode, Valid: m.RoutingMode != ""},
		sql.NullString{String: m.AccountGroup, Valid: m.AccountGroup != ""},
		metaJSON, m.ID)
	return err
}

// Delete removes a merchant by ID.
func (r *MerchantRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM merchants WHERE id = $1`, id)
	return err
}

// List retrieves merchants with pagination.
func (r *MerchantRepo) List(ctx context.Context, offset, limit int) ([]*model.Merchant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, status, routing_mode, created_at, updated_at
		FROM merchants ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []*model.Merchant
	for rows.Next() {
		m := &model.Merchant{}
		var email, routing sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &email, &m.Status, &routing, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		if email.Valid {
			m.Email = email.String
		}
		if routing.Valid {
			m.RoutingMode = routing.String
		}
		merchants = append(merchants, m)
	}
	return merchants, rows.Err()
}

// FindByAPIKeyHash retrieves a merchant by the SHA-256 hash of an API key.
func (r *MerchantRepo) FindByAPIKeyHash(ctx context.Context, keyHash string) (*model.Merchant, *model.APIKey, error) {
	var keyID, merchantID, keyPrefix string
	var permissions []string
	var ipWhitelist []string
	var createdAt, expiresAt, lastUsedAt, revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, merchant_id, key_prefix, permissions, ip_whitelist,
			created_at, expires_at, last_used_at, revoked_at
		FROM merchant_api_keys WHERE key_hash = $1`, keyHash,
	).Scan(&keyID, &merchantID, &keyPrefix, &permissions, &ipWhitelist,
		&createdAt, &expiresAt, &lastUsedAt, &revokedAt)
	if err != nil {
		return nil, nil, err
	}

	merchant, err := r.FindByID(ctx, merchantID)
	if err != nil {
		return nil, nil, err
	}

	k := &model.APIKey{
		KeyID:    keyID,
		KeyHash:  keyHash,
		KeyPrefix: keyPrefix,
		Permissions: permissions,
		IPWhitelist:  ipWhitelist,
	}
	if createdAt.Valid {
		k.CreatedAt = createdAt.Time
	}
	if expiresAt.Valid {
		k.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		k.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		k.RevokedAt = &revokedAt.Time
	}

	return merchant, k, nil
}

// CreateAPIKey inserts a new API key for a merchant.
func (r *MerchantRepo) CreateAPIKey(ctx context.Context, merchantID string, k *model.APIKey) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO merchant_api_keys (merchant_id, key_hash, key_prefix, permissions, ip_whitelist, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at`,
		merchantID, k.KeyHash, k.KeyPrefix, pq.Array(k.Permissions), pq.Array(k.IPWhitelist), k.ExpiresAt,
	).Scan(&k.KeyID, &k.CreatedAt)
}

// RevokeAPIKey marks an API key as revoked.
func (r *MerchantRepo) RevokeAPIKey(ctx context.Context, keyID string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE merchant_api_keys SET revoked_at = $1 WHERE id = $2`, now, keyID)
	return err
}

// ListAPIKeys returns all active (non-revoked) API keys for a merchant.
func (r *MerchantRepo) ListAPIKeys(ctx context.Context, merchantID string) ([]*model.APIKey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, key_hash, key_prefix, permissions, ip_whitelist,
			created_at, expires_at, last_used_at, revoked_at
		FROM merchant_api_keys
		WHERE merchant_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*model.APIKey
	for rows.Next() {
		k := &model.APIKey{}
		var perms, ips []string
		var createdAt, expiresAt, lastUsedAt, revokedAt sql.NullTime
		if err := rows.Scan(&k.KeyID, &k.KeyHash, &k.KeyPrefix, pq.Array(&perms), pq.Array(&ips),
			&createdAt, &expiresAt, &lastUsedAt, &revokedAt); err != nil {
			return nil, err
		}
		k.Permissions = perms
		k.IPWhitelist = ips
		if createdAt.Valid {
			k.CreatedAt = createdAt.Time
		}
		if expiresAt.Valid {
			k.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			k.LastUsedAt = &lastUsedAt.Time
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp for an API key.
func (r *MerchantRepo) UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE merchant_api_keys SET last_used_at = NOW() WHERE id = $1`, keyID)
	return err
}
