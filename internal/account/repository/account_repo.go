package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/account/model"
)

type AccountRepo struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) FindByID(ctx context.Context, id string) (*model.Account, error) {
	a := &model.Account{}
	var limitJSON, currencies, countries []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, gateway, alias, status, encrypted_cred,
		       b_site_id, merchant_id, limit_config, weight, priority,
		       tags, supported_currencies, supported_countries,
		       created_at, updated_at
		FROM payment_accounts WHERE id = $1`, id).
		Scan(&a.ID, &a.Gateway, &a.Alias, &a.Status, &a.EncryptedCred,
			&a.BSiteID, &a.MerchantID, &limitJSON, &a.Weight, &a.Priority,
			pq.Array(&a.Tags), &currencies, &countries,
			&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(limitJSON, &a.LimitConfig); err != nil {
		log.Warn().Err(err).Str("account_id", id).Msg("failed to unmarshal limit config")
	}
	if err := json.Unmarshal(currencies, &a.SupportedCurrencies); err != nil {
		log.Warn().Err(err).Str("account_id", id).Msg("failed to unmarshal currencies")
	}
	if err := json.Unmarshal(countries, &a.SupportedCountries); err != nil {
		log.Warn().Err(err).Str("account_id", id).Msg("failed to unmarshal countries")
	}
	return a, nil
}

func (r *AccountRepo) FindOnline(ctx context.Context, gateway, merchantID string) ([]*model.Account, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, gateway, alias, status, encrypted_cred,
		       b_site_id, merchant_id, limit_config, weight, priority,
		       tags, supported_currencies, supported_countries,
		       created_at, updated_at
		FROM payment_accounts
		WHERE gateway = $1 AND (merchant_id = $2 OR $2 = '00000000-0000-0000-0000-000000000000') AND status = 'online'
		ORDER BY priority DESC, weight DESC`, gateway, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		a := &model.Account{}
		var lj, cc, co []byte
		if err := rows.Scan(&a.ID, &a.Gateway, &a.Alias, &a.Status, &a.EncryptedCred,
			&a.BSiteID, &a.MerchantID, &lj, &a.Weight, &a.Priority,
			pq.Array(&a.Tags), &cc, &co, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(lj, &a.LimitConfig)
		json.Unmarshal(cc, &a.SupportedCurrencies)
		json.Unmarshal(co, &a.SupportedCountries)
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *AccountRepo) Create(ctx context.Context, a *model.Account) error {
	lj, _ := json.Marshal(a.LimitConfig)
	cc, _ := json.Marshal(a.SupportedCurrencies)
	co, _ := json.Marshal(a.SupportedCountries)

	return r.db.QueryRowContext(ctx, `
		INSERT INTO payment_accounts
		(gateway, alias, b_site_id, merchant_id, encrypted_cred,
		 status, weight, priority, tags, limit_config,
		 supported_currencies, supported_countries)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at`,
		a.Gateway, a.Alias, a.BSiteID, a.MerchantID, a.EncryptedCred,
		a.Status, a.Weight, a.Priority, pq.Array(a.Tags), lj, cc, co,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *AccountRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payment_accounts SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id)
	return err
}

func (r *AccountRepo) UpdateWeight(ctx context.Context, id string, weight int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payment_accounts SET weight = $1, updated_at = $2 WHERE id = $3`,
		weight, time.Now(), id)
	return err
}

func (r *AccountRepo) BatchImport(ctx context.Context, accounts []*model.Account) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	count := 0
	for _, a := range accounts {
		lj, _ := json.Marshal(a.LimitConfig)
		cc, _ := json.Marshal(a.SupportedCurrencies)
		co, _ := json.Marshal(a.SupportedCountries)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO payment_accounts
			(gateway, alias, b_site_id, merchant_id, encrypted_cred,
			 status, weight, priority, tags, limit_config,
			 supported_currencies, supported_countries)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT DO NOTHING`,
			a.Gateway, a.Alias, a.BSiteID, a.MerchantID, a.EncryptedCred,
			a.Status, a.Weight, a.Priority, pq.Array(a.Tags), lj, cc, co)
		if err == nil {
			count++
		}
	}
	if err := tx.Commit(); err != nil {
		return count, err
	}
	return count, nil
}

func (r *AccountRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM payment_accounts WHERE id = $1`, id)
	return err
}
