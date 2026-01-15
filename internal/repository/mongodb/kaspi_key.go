package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/seller-assistant/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type KaspiKeyRepository struct {
	collection *mongo.Collection
}

func NewKaspiKeyRepository(db *Database) *KaspiKeyRepository {
	return &KaspiKeyRepository{
		collection: db.DB.Collection("kaspi_keys"),
	}
}

func (r *KaspiKeyRepository) Create(key *domain.KaspiKey) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key.CreatedAt = time.Now()
	key.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to create kaspi key: %w", err)
	}

	key.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *KaspiKeyRepository) GetByUserID(userID string) (*domain.KaspiKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var key domain.KaspiKey
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&key)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kaspi key: %w", err)
	}

	return &key, nil
}

func (r *KaspiKeyRepository) GetByID(id string) (*domain.KaspiKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid kaspi key ID: %w", err)
	}

	var key domain.KaspiKey
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&key)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kaspi key: %w", err)
	}

	return &key, nil
}

func (r *KaspiKeyRepository) GetAllActive() ([]domain.KaspiKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, fmt.Errorf("failed to get active keys: %w", err)
	}
	defer cursor.Close(ctx)

	var keys []domain.KaspiKey
	if err := cursor.All(ctx, &keys); err != nil {
		return nil, fmt.Errorf("failed to decode active keys: %w", err)
	}

	return keys, nil
}

func (r *KaspiKeyRepository) Update(key *domain.KaspiKey) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key.UpdatedAt = time.Now()

	oid, err := primitive.ObjectIDFromHex(key.ID)
	if err != nil {
		return fmt.Errorf("invalid key ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"api_key_encrypted":    key.APIKeyEncrypted,
			"api_secret_encrypted": key.APISecretEncrypted,
			"merchant_id":          key.MerchantID,
			"is_active":            key.IsActive,
			"updated_at":           key.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *KaspiKeyRepository) Delete(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}
