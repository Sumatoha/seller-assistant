package mongodb

import (
	"context"
	"fmt"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

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

	// Users indexes
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"telegram_id": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("users").Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Kaspi keys indexes
	kaspiIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"user_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"is_active": 1},
		},
	}
	if _, err := d.DB.Collection("kaspi_keys").Indexes().CreateMany(ctx, kaspiIndexes); err != nil {
		return fmt.Errorf("failed to create kaspi_keys indexes: %w", err)
	}

	// Products indexes
	productsIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1},
		},
		{
			Keys: map[string]interface{}{"days_of_stock": 1},
		},
		{
			Keys:    map[string]interface{}{"user_id": 1, "external_id": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("products").Indexes().CreateMany(ctx, productsIndexes); err != nil {
		return fmt.Errorf("failed to create products indexes: %w", err)
	}

	// Sales history indexes
	salesIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"product_id": 1},
		},
		{
			Keys: map[string]interface{}{"date": -1},
		},
		{
			Keys:    map[string]interface{}{"product_id": 1, "date": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("sales_history").Indexes().CreateMany(ctx, salesIndexes); err != nil {
		return fmt.Errorf("failed to create sales_history indexes: %w", err)
	}

	// Reviews indexes
	reviewsIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1},
		},
		{
			Keys: map[string]interface{}{"ai_response_sent": 1},
		},
		{
			Keys:    map[string]interface{}{"user_id": 1, "external_id": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := d.DB.Collection("reviews").Indexes().CreateMany(ctx, reviewsIndexes); err != nil {
		return fmt.Errorf("failed to create reviews indexes: %w", err)
	}

	// Low stock alerts indexes
	alertsIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1},
		},
		{
			Keys: map[string]interface{}{"product_id": 1},
		},
		{
			Keys: map[string]interface{}{"notified_at": -1},
		},
	}
	if _, err := d.DB.Collection("low_stock_alerts").Indexes().CreateMany(ctx, alertsIndexes); err != nil {
		return fmt.Errorf("failed to create low_stock_alerts indexes: %w", err)
	}

	return nil
}
