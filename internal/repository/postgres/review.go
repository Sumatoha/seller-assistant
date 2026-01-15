package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/seller-assistant/internal/domain"
)

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(review *domain.Review) error {
	query := `
		INSERT INTO reviews
		(user_id, product_id, marketplace_key_id, external_id, author_name,
		 rating, comment, language, ai_response, ai_response_sent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		review.UserID,
		review.ProductID,
		review.MarketplaceKeyID,
		review.ExternalID,
		review.AuthorName,
		review.Rating,
		review.Comment,
		review.Language,
		review.AIResponse,
		review.AIResponseSent,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}

func (r *ReviewRepository) Update(review *domain.Review) error {
	review.UpdatedAt = time.Now()
	query := `
		UPDATE reviews
		SET ai_response = $1, ai_response_sent = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(
		query,
		review.AIResponse,
		review.AIResponseSent,
		review.UpdatedAt,
		review.ID,
	)

	return err
}

func (r *ReviewRepository) UpsertReview(review *domain.Review) error {
	query := `
		INSERT INTO reviews
		(user_id, product_id, marketplace_key_id, external_id, author_name,
		 rating, comment, language, ai_response, ai_response_sent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (marketplace_key_id, external_id)
		DO UPDATE SET
			rating = EXCLUDED.rating,
			comment = EXCLUDED.comment,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		review.UserID,
		review.ProductID,
		review.MarketplaceKeyID,
		review.ExternalID,
		review.AuthorName,
		review.Rating,
		review.Comment,
		review.Language,
		review.AIResponse,
		review.AIResponseSent,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}

func (r *ReviewRepository) GetByID(id int64) (*domain.Review, error) {
	var review domain.Review
	query := `SELECT * FROM reviews WHERE id = $1`

	err := r.db.Get(&review, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

func (r *ReviewRepository) GetPendingReviews(userID int64) ([]domain.Review, error) {
	var reviews []domain.Review
	query := `
		SELECT * FROM reviews
		WHERE user_id = $1 AND ai_response_sent = false
		ORDER BY created_at DESC
		LIMIT 50
	`

	err := r.db.Select(&reviews, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepository) GetByUserID(userID int64, limit int) ([]domain.Review, error) {
	var reviews []domain.Review
	query := `
		SELECT * FROM reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	err := r.db.Select(&reviews, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, nil
}
