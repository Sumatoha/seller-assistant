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

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{
		collection: db.DB.Collection("users"),
	}
}

func (r *UserRepository) Create(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	oid, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"username":           user.Username,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
			"language_code":      user.LanguageCode,
			"auto_reply_enabled": user.AutoReplyEnabled,
			"updated_at":         user.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *UserRepository) ToggleAutoReply(userID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"auto_reply_enabled": enabled,
			"updated_at":         time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"telegram_id": userID}, update)
	return err
}

func (r *UserRepository) ToggleAutoDumping(userID int64, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"auto_dumping_enabled": enabled,
			"updated_at":           time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"telegram_id": userID}, update)
	return err
}
