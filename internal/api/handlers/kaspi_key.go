package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/crypto"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type KaspiKeyHandler struct {
	kaspiKeyRepo domain.KaspiKeyRepository
	encryptor    *crypto.Encryptor
	syncService  *service.KaspiSyncService
}

func NewKaspiKeyHandler(kaspiKeyRepo domain.KaspiKeyRepository, encryptor *crypto.Encryptor, syncService *service.KaspiSyncService) *KaspiKeyHandler {
	return &KaspiKeyHandler{
		kaspiKeyRepo: kaspiKeyRepo,
		encryptor:    encryptor,
		syncService:  syncService,
	}
}

// CreateKaspiKeyRequest represents request to add Kaspi key
type CreateKaspiKeyRequest struct {
	APIKey     string `json:"api_key" binding:"required"`
	MerchantID string `json:"merchant_id" binding:"required"`
}

// KaspiKeyResponse represents Kaspi key without sensitive data
type KaspiKeyResponse struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
}

// GetKey returns user's Kaspi key (without API key)
// GET /api/v1/kaspi-key
func (h *KaspiKeyHandler) GetKey(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	key, err := h.kaspiKeyRepo.GetByUserID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kaspi key not found"})
		return
	}

	response := KaspiKeyResponse{
		ID:         key.ID,
		MerchantID: key.MerchantID,
		IsActive:   key.IsActive,
		CreatedAt:  key.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, response)
}

// CreateKey creates or updates Kaspi key
// POST /api/v1/kaspi-key
func (h *KaspiKeyHandler) CreateKey(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	var req CreateKaspiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Encrypt API key
	encryptedKey, err := h.encryptor.Encrypt(req.APIKey)
	if err != nil {
		logger.Log.Error("Failed to encrypt API key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
		return
	}

	// Check if key already exists
	existingKey, err := h.kaspiKeyRepo.GetByUserID(telegramID)
	if err == nil {
		// Update existing key
		existingKey.APIKeyEncrypted = encryptedKey
		existingKey.MerchantID = req.MerchantID
		existingKey.IsActive = true

		if err := h.kaspiKeyRepo.Update(existingKey); err != nil {
			logger.Log.Error("Failed to update Kaspi key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Kaspi key"})
			return
		}

		response := KaspiKeyResponse{
			ID:         existingKey.ID,
			MerchantID: existingKey.MerchantID,
			IsActive:   existingKey.IsActive,
			CreatedAt:  existingKey.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Kaspi key updated successfully",
			"key":     response,
		})
		return
	}

	// Create new key
	key := &domain.KaspiKey{
		UserID:          telegramID,
		APIKeyEncrypted: encryptedKey,
		MerchantID:      req.MerchantID,
		IsActive:        true,
	}

	if err := h.kaspiKeyRepo.Create(key); err != nil {
		logger.Log.Error("Failed to create Kaspi key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Kaspi key"})
		return
	}

	// Fetch created key
	createdKey, err := h.kaspiKeyRepo.GetByUserID(telegramID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get created key"})
		return
	}

	response := KaspiKeyResponse{
		ID:         createdKey.ID,
		MerchantID: createdKey.MerchantID,
		IsActive:   createdKey.IsActive,
		CreatedAt:  createdKey.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kaspi key created successfully",
		"key":     response,
	})
}

// DeleteKey deletes user's Kaspi key
// DELETE /api/v1/kaspi-key
func (h *KaspiKeyHandler) DeleteKey(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	// Check if key exists
	_, err := h.kaspiKeyRepo.GetByUserID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kaspi key not found"})
		return
	}

	if err := h.kaspiKeyRepo.Delete(telegramID); err != nil {
		logger.Log.Error("Failed to delete Kaspi key", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Kaspi key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kaspi key deleted successfully"})
}

// SyncNow triggers manual synchronization with Kaspi API
// POST /api/v1/kaspi-key/sync
func (h *KaspiKeyHandler) SyncNow(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Get user's Kaspi key
	kaspiKey, err := h.kaspiKeyRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kaspi key not found. Please configure your Kaspi API key first."})
		return
	}

	logger.Log.Info("Manual sync triggered",
		zap.String("user_id", userID),
	)

	// Run sync for this specific user
	if err := h.syncService.SyncUserData(kaspiKey); err != nil {
		logger.Log.Error("Manual sync failed",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Synchronization failed",
			"details": err.Error(),
		})
		return
	}

	logger.Log.Info("Manual sync completed successfully",
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Synchronization completed successfully",
	})
}
