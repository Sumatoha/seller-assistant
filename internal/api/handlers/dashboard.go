package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	productRepo domain.ProductRepository
	reviewRepo  domain.ReviewRepository
}

func NewDashboardHandler(productRepo domain.ProductRepository, reviewRepo domain.ReviewRepository) *DashboardHandler {
	return &DashboardHandler{
		productRepo: productRepo,
		reviewRepo:  reviewRepo,
	}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalProducts       int     `json:"total_products"`
	LowStockCount       int     `json:"low_stock_count"`
	DumpingEnabledCount int     `json:"dumping_enabled_count"`
	TotalReviews        int     `json:"total_reviews"`
	PendingReplies      int     `json:"pending_replies"`
	AverageRating       float64 `json:"average_rating"`
	TotalInventoryValue float64 `json:"total_inventory_value"`
}

// GetStats returns dashboard statistics
// GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	telegramID := middleware.GetTelegramID(c)

	stats := DashboardStats{}

	// Get products
	products, err := h.productRepo.GetByUserID(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	stats.TotalProducts = len(products)

	// Calculate product stats
	for _, p := range products {
		if p.DaysOfStock <= 7 && p.DaysOfStock > 0 {
			stats.LowStockCount++
		}
		if p.AutoDumpingEnabled {
			stats.DumpingEnabledCount++
		}
		stats.TotalInventoryValue += p.Price * float64(p.CurrentStock)
	}

	// Get reviews
	reviews, err := h.reviewRepo.GetByUserID(telegramID, 100)
	if err != nil {
		logger.Log.Error("Failed to get reviews", zap.Error(err))
	} else {
		stats.TotalReviews = len(reviews)

		// Calculate review stats
		totalRating := 0
		for _, r := range reviews {
			totalRating += r.Rating
			if r.AIResponse == "" {
				stats.PendingReplies++
			}
		}

		if len(reviews) > 0 {
			stats.AverageRating = float64(totalRating) / float64(len(reviews))
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetOverview returns dashboard overview with detailed data
// GET /api/v1/dashboard/overview
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	telegramID := middleware.GetTelegramID(c)

	// Get products
	products, err := h.productRepo.GetByUserID(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overview"})
		return
	}

	// Get low stock products
	lowStockProducts, err := h.productRepo.GetLowStockProducts(telegramID, 7)
	if err != nil {
		logger.Log.Error("Failed to get low stock products", zap.Error(err))
		lowStockProducts = []domain.Product{}
	}

	// Get dumping products
	dumpingProducts, err := h.productRepo.GetProductsForDumping(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get dumping products", zap.Error(err))
		dumpingProducts = []domain.Product{}
	}

	// Get recent reviews
	reviews, err := h.reviewRepo.GetByUserID(telegramID, 10)
	if err != nil {
		logger.Log.Error("Failed to get reviews", zap.Error(err))
		reviews = []domain.Review{}
	}

	// Get pending reviews
	pendingReviews, err := h.reviewRepo.GetPendingReviews(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get pending reviews", zap.Error(err))
		pendingReviews = []domain.Review{}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_products":    len(products),
		"low_stock":         lowStockProducts,
		"dumping_products":  dumpingProducts,
		"recent_reviews":    reviews,
		"pending_reviews":   pendingReviews,
		"low_stock_count":   len(lowStockProducts),
		"dumping_count":     len(dumpingProducts),
		"pending_count":     len(pendingReviews),
	})
}
