package service

import (
	"fmt"
	"time"

	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/marketplace/kaspi"
	"github.com/yourusername/seller-assistant/pkg/crypto"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type KaspiSyncService struct {
	kaspiKeyRepo     domain.KaspiKeyRepository
	productRepo      domain.ProductRepository
	salesHistoryRepo domain.SalesHistoryRepository
	reviewRepo       domain.ReviewRepository
	encryptor        *crypto.Encryptor
	inventoryService *InventoryService
}

func NewKaspiSyncService(
	kaspiKeyRepo domain.KaspiKeyRepository,
	productRepo domain.ProductRepository,
	salesHistoryRepo domain.SalesHistoryRepository,
	reviewRepo domain.ReviewRepository,
	encryptor *crypto.Encryptor,
	inventoryService *InventoryService,
) *KaspiSyncService {
	return &KaspiSyncService{
		kaspiKeyRepo:     kaspiKeyRepo,
		productRepo:      productRepo,
		salesHistoryRepo: salesHistoryRepo,
		reviewRepo:       reviewRepo,
		encryptor:        encryptor,
		inventoryService: inventoryService,
	}
}

// SyncAll syncs data for all active Kaspi keys
func (s *KaspiSyncService) SyncAll() error {
	keys, err := s.kaspiKeyRepo.GetAllActive()
	if err != nil {
		return fmt.Errorf("failed to get active keys: %w", err)
	}

	logger.Log.Info("Starting Kaspi sync", zap.Int("keys_count", len(keys)))

	for _, key := range keys {
		if err := s.SyncUserData(&key); err != nil {
			logger.Log.Error("Failed to sync user data",
				zap.Int64("user_id", key.UserID),
				zap.Error(err),
			)
			// Continue with other users
		}
	}

	logger.Log.Info("Kaspi sync completed")
	return nil
}

// SyncUserData syncs data for a specific user
func (s *KaspiSyncService) SyncUserData(key *domain.KaspiKey) error {
	client, err := s.getKaspiClient(key)
	if err != nil {
		return err
	}

	// Sync products
	if err := s.syncProducts(key.UserID, client); err != nil {
		logger.Log.Error("Failed to sync products", zap.Error(err))
	}

	// Sync sales data (last 7 days)
	if err := s.syncSalesData(key.UserID, client); err != nil {
		logger.Log.Error("Failed to sync sales data", zap.Error(err))
	}

	// Sync reviews
	if err := s.syncReviews(key.UserID, client); err != nil {
		logger.Log.Error("Failed to sync reviews", zap.Error(err))
	}

	// Recalculate inventory metrics
	if err := s.inventoryService.RecalculateAllProducts(key.UserID); err != nil {
		logger.Log.Error("Failed to recalculate inventory", zap.Error(err))
	}

	logger.Log.Info("User data synced successfully",
		zap.Int64("user_id", key.UserID),
	)

	return nil
}

func (s *KaspiSyncService) syncProducts(userID int64, client *kaspi.Client) error {
	products, err := client.GetProducts()
	if err != nil {
		return fmt.Errorf("failed to fetch products: %w", err)
	}

	logger.Log.Info("Syncing products",
		zap.Int64("user_id", userID),
		zap.Int("count", len(products)),
	)

	for _, p := range products {
		product := &domain.Product{
			UserID:       userID,
			ExternalID:   p.ExternalID,
			SKU:          p.SKU,
			Name:         p.Name,
			CurrentStock: p.CurrentStock,
			Price:        p.Price,
			Currency:     p.Currency,
			LastSyncAt:   time.Now(),
		}

		if err := s.productRepo.UpsertProduct(product); err != nil {
			logger.Log.Error("Failed to upsert product",
				zap.String("external_id", p.ExternalID),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *KaspiSyncService) syncSalesData(userID int64, client *kaspi.Client) error {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7) // Last 7 days

	salesData, err := client.GetSalesData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch sales data: %w", err)
	}

	logger.Log.Info("Syncing sales data",
		zap.Int64("user_id", userID),
		zap.Int("count", len(salesData)),
	)

	// Get all products for this user
	products, err := s.productRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	// Create a map of external ID to product ID
	productIDMap := make(map[string]string)
	for _, p := range products {
		productIDMap[p.ExternalID] = p.ID
	}

	// Group sales by product and date
	salesMap := make(map[string]map[string]*domain.SalesHistory)

	for _, sale := range salesData {
		dateKey := sale.Date.Format("2006-01-02")

		if _, ok := salesMap[sale.ProductExternalID]; !ok {
			salesMap[sale.ProductExternalID] = make(map[string]*domain.SalesHistory)
		}

		if _, ok := salesMap[sale.ProductExternalID][dateKey]; !ok {
			salesMap[sale.ProductExternalID][dateKey] = &domain.SalesHistory{
				Date:         sale.Date,
				QuantitySold: 0,
				Revenue:      0,
			}
		}

		salesMap[sale.ProductExternalID][dateKey].QuantitySold += sale.QuantitySold
		salesMap[sale.ProductExternalID][dateKey].Revenue += sale.Revenue
	}

	// Save sales history
	for externalID, dateMap := range salesMap {
		productID, ok := productIDMap[externalID]
		if !ok {
			continue
		}

		for _, history := range dateMap {
			history.ProductID = productID

			if err := s.salesHistoryRepo.UpsertSalesHistory(history); err != nil {
				logger.Log.Error("Failed to upsert sales history",
					zap.String("product_id", productID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

func (s *KaspiSyncService) syncReviews(userID int64, client *kaspi.Client) error {
	reviews, err := client.GetReviews()
	if err != nil {
		return fmt.Errorf("failed to fetch reviews: %w", err)
	}

	logger.Log.Info("Syncing reviews",
		zap.Int64("user_id", userID),
		zap.Int("count", len(reviews)),
	)

	// Get all products for this user
	products, err := s.productRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	// Create a map of external ID to product ID
	productIDMap := make(map[string]string)
	for _, p := range products {
		productIDMap[p.ExternalID] = p.ID
	}

	for _, r := range reviews {
		productID := ""
		if pid, ok := productIDMap[r.ProductID]; ok {
			productID = pid
		}

		review := &domain.Review{
			UserID:         userID,
			ProductID:      productID,
			ExternalID:     r.ExternalID,
			AuthorName:     r.AuthorName,
			Rating:         r.Rating,
			Comment:        r.Comment,
			Language:       r.Language,
			AIResponseSent: false,
		}

		if err := s.reviewRepo.UpsertReview(review); err != nil {
			logger.Log.Error("Failed to upsert review",
				zap.String("external_id", r.ExternalID),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *KaspiSyncService) getKaspiClient(key *domain.KaspiKey) (*kaspi.Client, error) {
	apiKey, err := s.encryptor.Decrypt(key.APIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return kaspi.NewClient(apiKey, key.MerchantID), nil
}
