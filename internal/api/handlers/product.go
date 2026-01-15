package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type ProductHandler struct {
	productRepo         domain.ProductRepository
	priceDumpingService *service.PriceDumpingService
}

func NewProductHandler(productRepo domain.ProductRepository, priceDumpingService *service.PriceDumpingService) *ProductHandler {
	return &ProductHandler{
		productRepo:         productRepo,
		priceDumpingService: priceDumpingService,
	}
}

// GetProducts returns all user's products
// GET /api/v1/products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	products, err := h.productRepo.GetByUserID(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// GetProduct returns single product by ID
// GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	productID := c.Param("id")

	product, err := h.productRepo.GetByID(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Verify ownership
	if product.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetLowStockProducts returns products with low stock
// GET /api/v1/products/low-stock
func (h *ProductHandler) GetLowStockProducts(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	products, err := h.productRepo.GetLowStockProducts(telegramID, 7)
	if err != nil {
		logger.Log.Error("Failed to get low stock products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get low stock products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// TEMPORARILY DISABLED - Price Dumping Feature
/*
// EnableDumpingRequest represents request to enable price dumping
type EnableDumpingRequest struct {
	MinPrice float64 `json:"min_price" binding:"required,gt=0"`
}

// EnableDumping enables price dumping for a product
// POST /api/v1/products/:id/dumping/enable
func (h *ProductHandler) EnableDumping(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	productID := c.Param("id")

	var req EnableDumpingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Verify product exists and ownership
	product, err := h.productRepo.GetByID(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Enable dumping
	if err := h.priceDumpingService.EnableProductDumping(productID, req.MinPrice); err != nil {
		logger.Log.Error("Failed to enable dumping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable dumping"})
		return
	}

	// Get updated product
	product, _ = h.productRepo.GetByID(productID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Price dumping enabled successfully",
		"product": product,
	})
}

// DisableDumping disables price dumping for a product
// POST /api/v1/products/:id/dumping/disable
func (h *ProductHandler) DisableDumping(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	productID := c.Param("id")

	// Verify product exists and ownership
	product, err := h.productRepo.GetByID(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if product.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Disable dumping
	if err := h.priceDumpingService.DisableProductDumping(productID); err != nil {
		logger.Log.Error("Failed to disable dumping", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable dumping"})
		return
	}

	// Get updated product
	product, _ = h.productRepo.GetByID(productID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Price dumping disabled successfully",
		"product": product,
	})
}

// GetDumpingProducts returns products with dumping enabled
// GET /api/v1/products/dumping
func (h *ProductHandler) GetDumpingProducts(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	products, err := h.productRepo.GetProductsForDumping(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get dumping products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dumping products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}
*/
