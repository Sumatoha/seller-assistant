package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/seller-assistant/internal/domain"
)

type MarketplaceKeyRepository struct {
	db *sqlx.DB
}

func NewMarketplaceKeyRepository(db *sqlx.DB) *MarketplaceKeyRepository {
	return &MarketplaceKeyRepository{db: db}
}

func (r *MarketplaceKeyRepository) Create(key *domain.MarketplaceKey) error {
	query := `
		INSERT INTO marketplace_keys
		(user_id, marketplace_type, api_key_encrypted, api_secret_encrypted, merchant_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		key.UserID,
		key.MarketplaceType,
		key.APIKeyEncrypted,
		key.APISecretEncrypted,
		key.MerchantID,
		key.IsActive,
	).Scan(&key.ID, &key.CreatedAt, &key.UpdatedAt)
}

func (r *MarketplaceKeyRepository) GetByUserID(userID int64) ([]domain.MarketplaceKey, error) {
	var keys []domain.MarketplaceKey
	query := `SELECT * FROM marketplace_keys WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.Select(&keys, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace keys: %w", err)
	}

	return keys, nil
}

func (r *MarketplaceKeyRepository) GetByID(id int64) (*domain.MarketplaceKey, error) {
	var key domain.MarketplaceKey
	query := `SELECT * FROM marketplace_keys WHERE id = $1`

	err := r.db.Get(&key, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace key: %w", err)
	}

	return &key, nil
}

func (r *MarketplaceKeyRepository) GetActiveKeys() ([]domain.MarketplaceKey, error) {
	var keys []domain.MarketplaceKey
	query := `SELECT * FROM marketplace_keys WHERE is_active = true`

	err := r.db.Select(&keys, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active keys: %w", err)
	}

	return keys, nil
}

func (r *MarketplaceKeyRepository) Update(key *domain.MarketplaceKey) error {
	key.UpdatedAt = time.Now()
	query := `
		UPDATE marketplace_keys
		SET api_key_encrypted = $1, api_secret_encrypted = $2,
		    merchant_id = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(
		query,
		key.APIKeyEncrypted,
		key.APISecretEncrypted,
		key.MerchantID,
		key.IsActive,
		key.UpdatedAt,
		key.ID,
	)

	return err
}

func (r *MarketplaceKeyRepository) Delete(id int64) error {
	query := `DELETE FROM marketplace_keys WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
