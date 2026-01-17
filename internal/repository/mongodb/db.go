package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Database struct {
	Client   *mongo.Client
	DB       *mongo.Database
	Database string
}

func NewDB(mongoURI, dbName string) (*Database, error) {
	// Increase timeout for cloud MongoDB (like Atlas)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Add timeout settings to client options
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetConnectTimeout(20 * time.Second).
		SetServerSelectionTimeout(20 * time.Second)

	fmt.Printf("[MongoDB] Connecting to MongoDB (database: %s)...\n", dbName)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("[MongoDB] ERROR: Connection failed: %v\n", err)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	fmt.Println("[MongoDB] Successfully connected, pinging database...")
	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Printf("[MongoDB] ERROR: Ping failed: %v\n", err)
		fmt.Printf("[MongoDB] ERROR: Error type: %T\n", err)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	fmt.Println("[MongoDB] Ping successful!")

	db := client.Database(dbName)

	database := &Database{
		Client:   client,
		DB:       db,
		Database: dbName,
	}

	// Create indexes
	if err := database.CreateIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return database, nil
}

func (d *Database) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return d.Client.Disconnect(ctx)
}

func (d *Database) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Users indexes - email is already indexed in user repository
	// No additional indexes needed here as email index is created in EnsureIndexes()

	// Kaspi keys indexes
	kaspiIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
	}
	if _, err := d.DB.Collection("kaspi_keys").Indexes().CreateMany(ctx, kaspiIndexes); err != nil {
		return fmt.Errorf("failed to create kaspi_keys indexes: %w", err)
	}

	// Products indexes
	productsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "days_of_stock", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "external_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("products").Indexes().CreateMany(ctx, productsIndexes); err != nil {
		return fmt.Errorf("failed to create products indexes: %w", err)
	}

	// Sales history indexes
	salesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "product_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "date", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("sales_history").Indexes().CreateMany(ctx, salesIndexes); err != nil {
		return fmt.Errorf("failed to create sales_history indexes: %w", err)
	}

	// Reviews indexes
	reviewsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "ai_response_sent", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "external_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("reviews").Indexes().CreateMany(ctx, reviewsIndexes); err != nil {
		return fmt.Errorf("failed to create reviews indexes: %w", err)
	}

	// Low stock alerts indexes
	alertsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "product_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "notified_at", Value: -1}},
		},
	}
	if _, err := d.DB.Collection("low_stock_alerts").Indexes().CreateMany(ctx, alertsIndexes); err != nil {
		return fmt.Errorf("failed to create low_stock_alerts indexes: %w", err)
	}

	return nil
}
