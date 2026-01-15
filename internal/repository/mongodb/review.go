package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/seller-assistant/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *Database) *ReviewRepository {
	return &ReviewRepository{
		collection: db.DB.Collection("reviews"),
	}
}

func (r *ReviewRepository) Create(review *domain.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	review.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *ReviewRepository) Update(review *domain.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	review.UpdatedAt = time.Now()

	oid, err := primitive.ObjectIDFromHex(review.ID)
	if err != nil {
		return fmt.Errorf("invalid review ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"ai_response":      review.AIResponse,
			"ai_response_sent": review.AIResponseSent,
			"updated_at":       review.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *ReviewRepository) UpsertReview(review *domain.Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	review.UpdatedAt = now
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}

	filter := bson.M{
		"user_id":     review.UserID,
		"external_id": review.ExternalID,
	}

	update := bson.M{
		"$set": bson.M{
			"user_id":     review.UserID,
			"product_id":  review.ProductID,
			"external_id": review.ExternalID,
			"author_name": review.AuthorName,
			"rating":      review.Rating,
			"comment":     review.Comment,
			"language":    review.Language,
			"updated_at":  review.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"ai_response":      review.AIResponse,
			"ai_response_sent": review.AIResponseSent,
			"created_at":       review.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert review: %w", err)
	}

	if result.UpsertedID != nil {
		review.ID = result.UpsertedID.(primitive.ObjectID).Hex()
	}

	return nil
}

func (r *ReviewRepository) GetByID(id string) (*domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid review ID: %w", err)
	}

	var review domain.Review
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&review)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

func (r *ReviewRepository) GetPendingReviews(userID string) ([]domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":          userID,
		"ai_response_sent": false,
	}

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(50)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}
	defer cursor.Close(ctx)

	var reviews []domain.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}

	return reviews, nil
}

func (r *ReviewRepository) GetByUserID(userID string, limit int) ([]domain.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer cursor.Close(ctx)

	var reviews []domain.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}

	return reviews, nil
}
