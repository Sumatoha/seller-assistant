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

const (
	// PriceDumpMargin - на сколько тенге ставим цену дешевле конкурентов
	PriceDumpMargin = 1.0
)

type PriceDumpingService struct {
	kaspiKeyRepo domain.KaspiKeyRepository
	productRepo  domain.ProductRepository
	encryptor    *crypto.Encryptor
}

func NewPriceDumpingService(
	kaspiKeyRepo domain.KaspiKeyRepository,
	productRepo domain.ProductRepository,
	encryptor *crypto.Encryptor,
) *PriceDumpingService {
	return &PriceDumpingService{
		kaspiKeyRepo: kaspiKeyRepo,
		productRepo:  productRepo,
		encryptor:    encryptor,
	}
}

// ProcessAllUsers обрабатывает автодемпинг для всех пользователей с включенной опцией
func (s *PriceDumpingService) ProcessAllUsers() error {
	keys, err := s.kaspiKeyRepo.GetAllActive()
	if err != nil {
		return fmt.Errorf("failed to get active keys: %w", err)
	}

	logger.Log.Info("Starting price dumping cycle", zap.Int("users_count", len(keys)))

	successCount := 0
	errorCount := 0

	for _, key := range keys {
		if err := s.ProcessUserProducts(key.UserID, &key); err != nil {
			logger.Log.Error("Failed to process user products",
				zap.String("user_id", key.UserID),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	logger.Log.Info("Price dumping cycle completed",
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// ProcessUserProducts обрабатывает автодемпинг для товаров конкретного пользователя
func (s *PriceDumpingService) ProcessUserProducts(userID string, key *domain.KaspiKey) error {
	// Получаем товары для демпинга
	products, err := s.productRepo.GetProductsForDumping(userID)
	if err != nil {
		return fmt.Errorf("failed to get products for dumping: %w", err)
	}

	if len(products) == 0 {
		logger.Log.Debug("No products for dumping", zap.String("user_id", userID))
		return nil
	}

	logger.Log.Info("Processing products for dumping",
		zap.String("user_id", userID),
		zap.Int("products_count", len(products)),
	)

	// Создаем Kaspi клиент
	client, err := s.getKaspiClient(key)
	if err != nil {
		return fmt.Errorf("failed to create Kaspi client: %w", err)
	}

	processedCount := 0
	updatedCount := 0

	for _, product := range products {
		if err := s.processProduct(&product, client); err != nil {
			logger.Log.Error("Failed to process product",
				zap.String("product_id", product.ID),
				zap.String("product_name", product.Name),
				zap.Error(err),
			)
			continue
		}
		processedCount++

		// Если цена была обновлена
		if product.LastPriceCheckAt.After(time.Now().Add(-10 * time.Second)) {
			updatedCount++
		}
	}

	logger.Log.Info("User products processed",
		zap.String("user_id", userID),
		zap.Int("processed", processedCount),
		zap.Int("updated", updatedCount),
	)

	return nil
}

// processProduct обрабатывает один товар
func (s *PriceDumpingService) processProduct(product *domain.Product, client *kaspi.Client) error {
	// Получаем цены конкурентов
	competitorPrices, err := client.GetCompetitorPrices(product.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to get competitor prices: %w", err)
	}

	if len(competitorPrices) == 0 {
		logger.Log.Debug("No competitors found", zap.String("product_id", product.ID))
		return nil
	}

	// Находим минимальную цену конкурента
	minCompetitorPrice := kaspi.GetMinCompetitorPrice(competitorPrices)

	// Вычисляем новую цену (на 1 тенге дешевле)
	newPrice := minCompetitorPrice - PriceDumpMargin

	// Проверяем минимальный порог
	if product.MinPrice > 0 && newPrice < product.MinPrice {
		logger.Log.Info("Price below minimum threshold, skipping",
			zap.String("product_id", product.ID),
			zap.String("product_name", product.Name),
			zap.Float64("new_price", newPrice),
			zap.Float64("min_price", product.MinPrice),
			zap.Float64("competitor_price", minCompetitorPrice),
		)

		// Обновляем только информацию о цене конкурента
		if err := s.productRepo.UpdatePrice(product.ID, product.Price, minCompetitorPrice); err != nil {
			return fmt.Errorf("failed to update competitor price: %w", err)
		}

		return nil
	}

	// Проверяем, нужно ли менять цену
	if product.Price == newPrice {
		logger.Log.Debug("Price already optimal",
			zap.String("product_id", product.ID),
			zap.Float64("current_price", product.Price),
		)

		// Обновляем время проверки и цену конкурента
		if err := s.productRepo.UpdatePrice(product.ID, product.Price, minCompetitorPrice); err != nil {
			return fmt.Errorf("failed to update price check time: %w", err)
		}

		return nil
	}

	// Обновляем цену на Kaspi
	if err := client.UpdateProductPrice(product.ExternalID, newPrice); err != nil {
		return fmt.Errorf("failed to update price on Kaspi: %w", err)
	}

	// Обновляем цену в БД
	if err := s.productRepo.UpdatePrice(product.ID, newPrice, minCompetitorPrice); err != nil {
		return fmt.Errorf("failed to update price in database: %w", err)
	}

	logger.Log.Info("Price updated successfully",
		zap.String("product_id", product.ID),
		zap.String("product_name", product.Name),
		zap.Float64("old_price", product.Price),
		zap.Float64("new_price", newPrice),
		zap.Float64("min_competitor_price", minCompetitorPrice),
		zap.Float64("min_threshold", product.MinPrice),
	)

	return nil
}

func (s *PriceDumpingService) getKaspiClient(key *domain.KaspiKey) (*kaspi.Client, error) {
	apiKey, err := s.encryptor.Decrypt(key.APIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return kaspi.NewClient(apiKey, key.MerchantID), nil
}

// EnableProductDumping включает автодемпинг для конкретного товара
func (s *PriceDumpingService) EnableProductDumping(productID string, minPrice float64) error {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return fmt.Errorf("product not found")
	}

	product.AutoDumpingEnabled = true
	product.MinPrice = minPrice

	if err := s.productRepo.Update(product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	logger.Log.Info("Auto dumping enabled for product",
		zap.String("product_id", productID),
		zap.String("product_name", product.Name),
		zap.Float64("min_price", minPrice),
	)

	return nil
}

// DisableProductDumping выключает автодемпинг для конкретного товара
func (s *PriceDumpingService) DisableProductDumping(productID string) error {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return fmt.Errorf("product not found")
	}

	product.AutoDumpingEnabled = false

	if err := s.productRepo.Update(product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	logger.Log.Info("Auto dumping disabled for product",
		zap.String("product_id", productID),
		zap.String("product_name", product.Name),
	)

	return nil
}
