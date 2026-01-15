package service

import (
	"fmt"
	"math"
	"time"

	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type InventoryService struct {
	productRepo      domain.ProductRepository
	salesHistoryRepo domain.SalesHistoryRepository
	alertRepo        domain.LowStockAlertRepository
}

func NewInventoryService(
	productRepo domain.ProductRepository,
	salesHistoryRepo domain.SalesHistoryRepository,
	alertRepo domain.LowStockAlertRepository,
) *InventoryService {
	return &InventoryService{
		productRepo:      productRepo,
		salesHistoryRepo: salesHistoryRepo,
		alertRepo:        alertRepo,
	}
}

// CalculateDaysOfStock calculates how many days of stock remain based on sales velocity
func (s *InventoryService) CalculateDaysOfStock(productID string) (int, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return 0, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return 0, fmt.Errorf("product not found")
	}

	// Get sales history for the last 30 days
	salesHistory, err := s.salesHistoryRepo.GetByProductID(productID, 30)
	if err != nil {
		return 0, fmt.Errorf("failed to get sales history: %w", err)
	}

	// Calculate average daily sales (sales velocity)
	salesVelocity := s.calculateSalesVelocity(salesHistory)

	// Calculate days of stock
	daysOfStock := 0
	if salesVelocity > 0 {
		daysOfStock = int(math.Ceil(float64(product.CurrentStock) / salesVelocity))
	} else if product.CurrentStock > 0 {
		// If no sales history, assume infinite stock
		daysOfStock = 999
	}

	// Update product with new calculations
	product.SalesVelocity = salesVelocity
	product.DaysOfStock = daysOfStock
	product.LastSyncAt = time.Now()

	if err := s.productRepo.Update(product); err != nil {
		return 0, fmt.Errorf("failed to update product: %w", err)
	}

	return daysOfStock, nil
}

// calculateSalesVelocity calculates average daily sales
func (s *InventoryService) calculateSalesVelocity(salesHistory []domain.SalesHistory) float64 {
	if len(salesHistory) == 0 {
		return 0
	}

	totalSold := 0
	for _, sale := range salesHistory {
		totalSold += sale.QuantitySold
	}

	// Use actual number of days with data
	daysWithData := len(salesHistory)
	if daysWithData == 0 {
		return 0
	}

	return float64(totalSold) / float64(daysWithData)
}

// ProcessLowStockAlerts checks for low stock and creates alerts
func (s *InventoryService) ProcessLowStockAlerts(userID int64, thresholdDays int) error {
	products, err := s.productRepo.GetLowStockProducts(userID, thresholdDays)
	if err != nil {
		return fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Check if we've already sent recent alerts (within last 24 hours)
	recentAlerts, err := s.alertRepo.GetRecentAlerts(userID, 24)
	if err != nil {
		return fmt.Errorf("failed to get recent alerts: %w", err)
	}

	// Create map of products with recent alerts
	alertedProducts := make(map[string]bool)
	for _, alert := range recentAlerts {
		alertedProducts[alert.ProductID] = true
	}

	// Create new alerts for products without recent alerts
	for _, product := range products {
		if !alertedProducts[product.ID] {
			alert := &domain.LowStockAlert{
				ProductID:     product.ID,
				UserID:        userID,
				ThresholdDays: thresholdDays,
			}

			if err := s.alertRepo.Create(alert); err != nil {
				logger.Log.Error("Failed to create low stock alert",
					zap.String("product_id", product.ID),
					zap.Error(err),
				)
				continue
			}

			logger.Log.Info("Created low stock alert",
				zap.Int64("user_id", userID),
				zap.String("product_id", product.ID),
				zap.String("product_name", product.Name),
				zap.Int("days_of_stock", product.DaysOfStock),
			)
		}
	}

	return nil
}

// GetLowStockSummary returns a summary of low stock products
func (s *InventoryService) GetLowStockSummary(userID int64, thresholdDays int) ([]domain.Product, error) {
	products, err := s.productRepo.GetLowStockProducts(userID, thresholdDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	return products, nil
}

// RecalculateAllProducts recalculates days of stock for all user products
func (s *InventoryService) RecalculateAllProducts(userID int64) error {
	products, err := s.productRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	for _, product := range products {
		if _, err := s.CalculateDaysOfStock(product.ID); err != nil {
			logger.Log.Error("Failed to calculate days of stock",
				zap.String("product_id", product.ID),
				zap.Error(err),
			)
			// Continue with other products even if one fails
		}
	}

	return nil
}
