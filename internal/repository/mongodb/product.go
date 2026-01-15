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

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *Database) *ProductRepository {
	return &ProductRepository{
		collection: db.DB.Collection("products"),
	}
}

func (r *ProductRepository) Create(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	product.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *ProductRepository) Update(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product.UpdatedAt = time.Now()

	oid, err := primitive.ObjectIDFromHex(product.ID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"current_stock":        product.CurrentStock,
			"price":                product.Price,
			"min_price":            product.MinPrice,
			"competitor_min_price": product.CompetitorMinPrice,
			"auto_dumping_enabled": product.AutoDumpingEnabled,
			"sales_velocity":       product.SalesVelocity,
			"days_of_stock":        product.DaysOfStock,
			"last_price_check_at":  product.LastPriceCheckAt,
			"last_sync_at":         product.LastSyncAt,
			"updated_at":           product.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *ProductRepository) UpdatePrice(id string, newPrice float64, competitorMinPrice float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"price":                newPrice,
			"competitor_min_price": competitorMinPrice,
			"last_price_check_at":  time.Now(),
			"updated_at":           time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

func (r *ProductRepository) GetProductsForDumping(userID int64) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":              userID,
		"auto_dumping_enabled": true,
		"current_stock": bson.M{
			"$gt": 0, // Только товары в наличии
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get products for dumping: %w", err)
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) UpsertProduct(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	product.UpdatedAt = now
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}

	filter := bson.M{
		"user_id":     product.UserID,
		"external_id": product.ExternalID,
	}

	update := bson.M{
		"$set": bson.M{
			"user_id":        product.UserID,
			"external_id":    product.ExternalID,
			"sku":            product.SKU,
			"name":           product.Name,
			"current_stock":  product.CurrentStock,
			"price":          product.Price,
			"currency":       product.Currency,
			"sales_velocity": product.SalesVelocity,
			"days_of_stock":  product.DaysOfStock,
			"last_sync_at":   product.LastSyncAt,
			"updated_at":     product.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": product.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert product: %w", err)
	}

	if result.UpsertedID != nil {
		product.ID = result.UpsertedID.(primitive.ObjectID).Hex()
	}

	return nil
}

func (r *ProductRepository) GetByID(id string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	var product domain.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&product)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetByUserID(userID int64) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{"days_of_stock", 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetLowStockProducts(userID int64, thresholdDays int) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"days_of_stock": bson.M{
			"$lte": thresholdDays,
			"$gt":  0,
		},
	}

	opts := options.Find().SetSort(bson.D{{"days_of_stock", 1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

// SalesHistoryRepository
type SalesHistoryRepository struct {
	collection *mongo.Collection
}

func NewSalesHistoryRepository(db *Database) *SalesHistoryRepository {
	return &SalesHistoryRepository{
		collection: db.DB.Collection("sales_history"),
	}
}

func (r *SalesHistoryRepository) Create(history *domain.SalesHistory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	history.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to create sales history: %w", err)
	}

	history.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *SalesHistoryRepository) UpsertSalesHistory(history *domain.SalesHistory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	if history.CreatedAt.IsZero() {
		history.CreatedAt = now
	}

	filter := bson.M{
		"product_id": history.ProductID,
		"date":       history.Date,
	}

	update := bson.M{
		"$set": bson.M{
			"product_id":    history.ProductID,
			"date":          history.Date,
			"quantity_sold": history.QuantitySold,
			"revenue":       history.Revenue,
		},
		"$setOnInsert": bson.M{
			"created_at": history.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert sales history: %w", err)
	}

	if result.UpsertedID != nil {
		history.ID = result.UpsertedID.(primitive.ObjectID).Hex()
	}

	return nil
}

func (r *SalesHistoryRepository) GetByProductID(productID string, days int) ([]domain.SalesHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"product_id": productID,
		"date": bson.M{
			"$gte": time.Now().AddDate(0, 0, -days),
		},
	}

	opts := options.Find().SetSort(bson.D{{"date", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []domain.SalesHistory
	if err := cursor.All(ctx, &history); err != nil {
		return nil, fmt.Errorf("failed to decode sales history: %w", err)
	}

	return history, nil
}

// LowStockAlertRepository
type LowStockAlertRepository struct {
	collection *mongo.Collection
}

func NewLowStockAlertRepository(db *Database) *LowStockAlertRepository {
	return &LowStockAlertRepository{
		collection: db.DB.Collection("low_stock_alerts"),
	}
}

func (r *LowStockAlertRepository) Create(alert *domain.LowStockAlert) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alert.NotifiedAt = time.Now()
	alert.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, alert)
	if err != nil {
		return fmt.Errorf("failed to create low stock alert: %w", err)
	}

	alert.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *LowStockAlertRepository) GetRecentAlerts(userID int64, hours int) ([]domain.LowStockAlert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"notified_at": bson.M{
			"$gte": time.Now().Add(-time.Duration(hours) * time.Hour),
		},
	}

	opts := options.Find().SetSort(bson.D{{"notified_at", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}
	defer cursor.Close(ctx)

	var alerts []domain.LowStockAlert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, fmt.Errorf("failed to decode alerts: %w", err)
	}

	return alerts, nil
}
