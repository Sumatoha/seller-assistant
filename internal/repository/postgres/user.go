package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/seller-assistant/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, language_code, auto_reply_enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		user.AutoReplyEnabled,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE telegram_id = $1`

	err := r.db.Get(&user, query, telegramID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3,
		    language_code = $4, auto_reply_enabled = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		user.AutoReplyEnabled,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

func (r *UserRepository) ToggleAutoReply(userID int64, enabled bool) error {
	query := `UPDATE users SET auto_reply_enabled = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, enabled, time.Now(), userID)
	return err
}
