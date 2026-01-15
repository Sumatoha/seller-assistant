package domain

import "time"

type Review struct {
	ID             string    `bson:"_id,omitempty" json:"id"`
	UserID         int64     `bson:"user_id" json:"user_id"`
	ProductID      string    `bson:"product_id,omitempty" json:"product_id,omitempty"` // Reference to Product._id
	ExternalID     string    `bson:"external_id" json:"external_id"`                   // Kaspi review ID
	AuthorName     string    `bson:"author_name" json:"author_name"`
	Rating         int       `bson:"rating" json:"rating"`
	Comment        string    `bson:"comment" json:"comment"`
	Language       string    `bson:"language" json:"language"`
	AIResponse     string    `bson:"ai_response" json:"ai_response"`
	AIResponseSent bool      `bson:"ai_response_sent" json:"ai_response_sent"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

type ReviewRepository interface {
	Create(review *Review) error
	Update(review *Review) error
	GetByID(id string) (*Review, error)
	GetPendingReviews(userID int64) ([]Review, error)
	GetByUserID(userID int64, limit int) ([]Review, error)
	UpsertReview(review *Review) error
}
