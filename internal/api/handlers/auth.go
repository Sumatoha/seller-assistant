package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/yourusername/seller-assistant/internal/api/middleware"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userRepo           domain.UserRepository
	jwtSecret          string
	jwtExpirationHours int
}

func NewAuthHandler(userRepo domain.UserRepository, jwtSecret string, jwtExpirationHours int) *AuthHandler {
	return &AuthHandler{
		userRepo:           userRepo,
		jwtSecret:          jwtSecret,
		jwtExpirationHours: jwtExpirationHours,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	Language  string `json:"language"` // "ru" or "kk"
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

// Register registers a new user
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user with this email already exists
	existingUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		logger.Log.Error("Failed to check existing user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// Set default language
	language := req.Language
	if language == "" {
		language = "ru"
	}

	// Create new user
	user := &domain.User{
		Email:              req.Email,
		PasswordHash:       string(passwordHash),
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		LanguageCode:       language,
		AutoReplyEnabled:   false,
		AutoDumpingEnabled: false,
	}

	if err := h.userRepo.Create(user); err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	logger.Log.Info("New user registered",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		logger.Log.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Login authenticates a user
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		logger.Log.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(user)
	if err != nil {
		logger.Log.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	logger.Log.Info("User logged in",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
	)

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  *user,
	})
}

// GetMe returns current user information
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		logger.Log.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// generateToken creates a JWT token for the user
func (h *AuthHandler) generateToken(user *domain.User) (string, error) {
	expirationHours := h.jwtExpirationHours
	if expirationHours == 0 {
		expirationHours = 168 // 7 days by default
	}

	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
