package domain

import "time"

type KaspiKey struct {
	ID                 string    `bson:"_id,omitempty" json:"id"`
	UserID             string    `bson:"user_id" json:"user_id"`
	APIKeyEncrypted    string    `bson:"api_key_encrypted" json:"-"`
	APISecretEncrypted string    `bson:"api_secret_encrypted" json:"-"`
	MerchantID         string    `bson:"merchant_id" json:"merchant_id"`
	IsActive           bool      `bson:"is_active" json:"is_active"`
	CreatedAt          time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at" json:"updated_at"`
}

type KaspiKeyRepository interface {
	Create(key *KaspiKey) error
	GetByUserID(userID string) (*KaspiKey, error)
	GetByID(id string) (*KaspiKey, error)
	GetAllActive() ([]KaspiKey, error)
	Update(key *KaspiKey) error
	Delete(userID string) error
}
