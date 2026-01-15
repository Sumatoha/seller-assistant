package domain

import "time"

type Product struct {
	ID                 string    `bson:"_id,omitempty" json:"id"`
	UserID             string    `bson:"user_id" json:"user_id"`
	ExternalID         string    `bson:"external_id" json:"external_id"` // Kaspi product ID
	SKU                string    `bson:"sku" json:"sku"`
	Name               string    `bson:"name" json:"name"`
	CurrentStock       int       `bson:"current_stock" json:"current_stock"`
	Price              float64   `bson:"price" json:"price"`
	MinPrice           float64   `bson:"min_price" json:"min_price"`                       // Минимальная цена для демпинга
	CompetitorMinPrice float64   `bson:"competitor_min_price" json:"competitor_min_price"` // Минимальная цена конкурентов
	AutoDumpingEnabled bool      `bson:"auto_dumping_enabled" json:"auto_dumping_enabled"` // Включен ли автодемпинг
	Currency           string    `bson:"currency" json:"currency"`
	SalesVelocity      float64   `bson:"sales_velocity" json:"sales_velocity"`
	DaysOfStock        int       `bson:"days_of_stock" json:"days_of_stock"`
	LastPriceCheckAt   time.Time `bson:"last_price_check_at" json:"last_price_check_at"`
	LastSyncAt         time.Time `bson:"last_sync_at" json:"last_sync_at"`
	CreatedAt          time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at" json:"updated_at"`
}

type SalesHistory struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	ProductID    string    `bson:"product_id" json:"product_id"`
	Date         time.Time `bson:"date" json:"date"`
	QuantitySold int       `bson:"quantity_sold" json:"quantity_sold"`
	Revenue      float64   `bson:"revenue" json:"revenue"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
}

type LowStockAlert struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	ProductID     string    `bson:"product_id" json:"product_id"`
	UserID        string    `bson:"user_id" json:"user_id"`
	ThresholdDays int       `bson:"threshold_days" json:"threshold_days"`
	NotifiedAt    time.Time `bson:"notified_at" json:"notified_at"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

type ProductRepository interface {
	Create(product *Product) error
	Update(product *Product) error
	UpdatePrice(id string, newPrice float64, competitorMinPrice float64) error
	GetByID(id string) (*Product, error)
	GetByUserID(userID string) ([]Product, error)
	GetProductsForDumping(userID string) ([]Product, error)
	GetLowStockProducts(userID string, thresholdDays int) ([]Product, error)
	UpsertProduct(product *Product) error
}

type SalesHistoryRepository interface {
	Create(history *SalesHistory) error
	GetByProductID(productID string, days int) ([]SalesHistory, error)
	UpsertSalesHistory(history *SalesHistory) error
}

type LowStockAlertRepository interface {
	Create(alert *LowStockAlert) error
	GetRecentAlerts(userID string, hours int) ([]LowStockAlert, error)
}
