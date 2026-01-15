package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userRepo domain.UserRepository
}

func NewAuthHandler(userRepo domain.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
	}
}

// LoginRequest represents login credentials
type LoginRequest struct {
	TelegramID int64  `json:"telegram_id" binding:"required"`
	Username   string `json:"username"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

// Login authenticates user and returns JWT token
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get or create user
	user, err := h.userRepo.GetByTelegramID(req.TelegramID)
	if err != nil {
		// User doesn't exist, create new one
		user = &domain.User{
			TelegramID:         req.TelegramID,
			Username:           req.Username,
			LanguageCode:       "ru",
			AutoReplyEnabled:   false,
			AutoDumpingEnabled: false,
		}

		if err := h.userRepo.Create(user); err != nil {
			logger.Log.Error("Failed to create user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Fetch the created user to get ID
		user, err = h.userRepo.GetByTelegramID(req.TelegramID)
		if err != nil {
			logger.Log.Error("Failed to get created user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
	}

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		logger.Log.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  *user,
	})
}

// generateToken creates a JWT token for the user
func (h *AuthHandler) generateToken(user *domain.User) (string, error) {
	claims := &middleware.Claims{
		UserID:     user.TelegramID, // Using TelegramID as UserID for simplicity
		TelegramID: user.TelegramID,
		Username:   user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret()))
}

// getJWTSecret retrieves JWT secret from environment or uses default
func getJWTSecret() string {
	// TODO: Get from config
	return "your-super-secret-jwt-key-change-in-production"
}

// GetMe returns current user info
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	telegramID := middleware.GetTelegramID(c)

	user, err := h.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
