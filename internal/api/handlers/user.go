package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type UserHandler struct {
	userRepo domain.UserRepository
}

func NewUserHandler(userRepo domain.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// UpdateSettingsRequest represents settings update request
type UpdateSettingsRequest struct {
	AutoReplyEnabled   *bool   `json:"auto_reply_enabled"`
	AutoDumpingEnabled *bool   `json:"auto_dumping_enabled"`
	Language           *string `json:"language"`
}

// GetProfile returns user profile
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateSettings updates user settings
// PATCH /api/v1/user/settings
func (h *UserHandler) UpdateSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update settings
	if req.AutoReplyEnabled != nil {
		if err := h.userRepo.ToggleAutoReply(userID, *req.AutoReplyEnabled); err != nil {
			logger.Log.Error("Failed to toggle auto-reply", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auto-reply"})
			return
		}
		user.AutoReplyEnabled = *req.AutoReplyEnabled
	}

	if req.AutoDumpingEnabled != nil {
		if err := h.userRepo.ToggleAutoDumping(userID, *req.AutoDumpingEnabled); err != nil {
			logger.Log.Error("Failed to toggle auto-dumping", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update auto-dumping"})
			return
		}
		user.AutoDumpingEnabled = *req.AutoDumpingEnabled
	}

	if req.Language != nil {
		user.LanguageCode = *req.Language
		if err := h.userRepo.Update(user); err != nil {
			logger.Log.Error("Failed to update language", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update language"})
			return
		}
	}

	// Return updated user
	user, err = h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
