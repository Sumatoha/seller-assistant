package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yourusername/seller-assistant/internal/domain"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *domain.Product) error {
	query := `
		INSERT INTO products
		(user_id, marketplace_key_id, external_id, sku, name, current_stock,
		 price, currency, sales_velocity, days_of_stock, last_sync_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		product.UserID,
		product.MarketplaceKeyID,
		product.ExternalID,
		product.SKU,
		product.Name,
		product.CurrentStock,
		product.Price,
		product.Currency,
		product.SalesVelocity,
		product.DaysOfStock,
		product.LastSyncAt,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) Update(product *domain.Product) error {
	product.UpdatedAt = time.Now()
	query := `
		UPDATE products
		SET current_stock = $1, price = $2, sales_velocity = $3,
		    days_of_stock = $4, last_sync_at = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(
		query,
		product.CurrentStock,
		product.Price,
		product.SalesVelocity,
		product.DaysOfStock,
		product.LastSyncAt,
		product.UpdatedAt,
		product.ID,
	)

	return err
}

func (r *ProductRepository) UpsertProduct(product *domain.Product) error {
	query := `
		INSERT INTO products
		(user_id, marketplace_key_id, external_id, sku, name, current_stock,
		 price, currency, sales_velocity, days_of_stock, last_sync_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (marketplace_key_id, external_id)
		DO UPDATE SET
			current_stock = EXCLUDED.current_stock,
			price = EXCLUDED.price,
			sales_velocity = EXCLUDED.sales_velocity,
			days_of_stock = EXCLUDED.days_of_stock,
			last_sync_at = EXCLUDED.last_sync_at,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		product.UserID,
		product.MarketplaceKeyID,
		product.ExternalID,
		product.SKU,
		product.Name,
		product.CurrentStock,
		product.Price,
		product.Currency,
		product.SalesVelocity,
		product.DaysOfStock,
		product.LastSyncAt,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetByID(id int64) (*domain.Product, error) {
	var product domain.Product
	query := `SELECT * FROM products WHERE id = $1`

	err := r.db.Get(&product, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetByUserID(userID int64) ([]domain.Product, error) {
	var products []domain.Product
	query := `SELECT * FROM products WHERE user_id = $1 ORDER BY days_of_stock ASC`

	err := r.db.Select(&products, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetByMarketplaceKeyID(keyID int64) ([]domain.Product, error) {
	var products []domain.Product
	query := `SELECT * FROM products WHERE marketplace_key_id = $1`

	err := r.db.Select(&products, query, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetLowStockProducts(userID int64, thresholdDays int) ([]domain.Product, error) {
	var products []domain.Product
	query := `
		SELECT * FROM products
		WHERE user_id = $1 AND days_of_stock <= $2 AND days_of_stock > 0
		ORDER BY days_of_stock ASC
	`

	err := r.db.Select(&products, query, userID, thresholdDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	return products, nil
}

// SalesHistoryRepository
type SalesHistoryRepository struct {
	db *sqlx.DB
}

func NewSalesHistoryRepository(db *sqlx.DB) *SalesHistoryRepository {
	return &SalesHistoryRepository{db: db}
}

func (r *SalesHistoryRepository) Create(history *domain.SalesHistory) error {
	query := `
		INSERT INTO sales_history (product_id, date, quantity_sold, revenue)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		history.ProductID,
		history.Date,
		history.QuantitySold,
		history.Revenue,
	).Scan(&history.ID, &history.CreatedAt)
}

func (r *SalesHistoryRepository) UpsertSalesHistory(history *domain.SalesHistory) error {
	query := `
		INSERT INTO sales_history (product_id, date, quantity_sold, revenue)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (product_id, date)
		DO UPDATE SET
			quantity_sold = EXCLUDED.quantity_sold,
			revenue = EXCLUDED.revenue
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		query,
		history.ProductID,
		history.Date,
		history.QuantitySold,
		history.Revenue,
	).Scan(&history.ID, &history.CreatedAt)
}

func (r *SalesHistoryRepository) GetByProductID(productID int64, days int) ([]domain.SalesHistory, error) {
	var history []domain.SalesHistory
	query := `
		SELECT * FROM sales_history
		WHERE product_id = $1 AND date >= NOW() - INTERVAL '%d days'
		ORDER BY date DESC
	`

	err := r.db.Select(&history, fmt.Sprintf(query, days), productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales history: %w", err)
	}

	return history, nil
}

// LowStockAlertRepository
type LowStockAlertRepository struct {
	db *sqlx.DB
}

func NewLowStockAlertRepository(db *sqlx.DB) *LowStockAlertRepository {
	return &LowStockAlertRepository{db: db}
}

func (r *LowStockAlertRepository) Create(alert *domain.LowStockAlert) error {
	query := `
		INSERT INTO low_stock_alerts (product_id, user_id, threshold_days)
		VALUES ($1, $2, $3)
		RETURNING id, notified_at, created_at
	`

	return r.db.QueryRow(
		query,
		alert.ProductID,
		alert.UserID,
		alert.ThresholdDays,
	).Scan(&alert.ID, &alert.NotifiedAt, &alert.CreatedAt)
}

func (r *LowStockAlertRepository) GetRecentAlerts(userID int64, hours int) ([]domain.LowStockAlert, error) {
	var alerts []domain.LowStockAlert
	query := `
		SELECT * FROM low_stock_alerts
		WHERE user_id = $1 AND notified_at >= NOW() - INTERVAL '%d hours'
		ORDER BY notified_at DESC
	`

	err := r.db.Select(&alerts, fmt.Sprintf(query, hours), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}

	return alerts, nil
}
