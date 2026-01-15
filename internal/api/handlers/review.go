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

type ReviewHandler struct {
	reviewRepo  domain.ReviewRepository
	aiResponder *service.AIResponderService
}

func NewReviewHandler(reviewRepo domain.ReviewRepository, aiResponder *service.AIResponderService) *ReviewHandler {
	return &ReviewHandler{
		reviewRepo:  reviewRepo,
		aiResponder: aiResponder,
	}
}

// GetReviews returns all user's reviews
// GET /api/v1/reviews
func (h *ReviewHandler) GetReviews(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	// Query params
	limit := 50
	if l := c.Query("limit"); l != "" {
		if _, err := c.GetQuery("limit"); err {
			limit = 50
		}
	}

	reviews, err := h.reviewRepo.GetByUserID(telegramID, limit)
	if err != nil {
		logger.Log.Error("Failed to get reviews", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"count":   len(reviews),
	})
}

// GetReview returns single review by ID
// GET /api/v1/reviews/:id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	reviewID := c.Param("id")

	review, err := h.reviewRepo.GetByID(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Verify ownership
	if review.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetPendingReviews returns reviews without AI response
// GET /api/v1/reviews/pending
func (h *ReviewHandler) GetPendingReviews(c *gin.Context) {
	telegramID := middleware.GetUserID(c)

	reviews, err := h.reviewRepo.GetPendingReviews(telegramID)
	if err != nil {
		logger.Log.Error("Failed to get pending reviews", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"count":   len(reviews),
	})
}

// GenerateReplyRequest represents request to generate AI reply
type GenerateReplyRequest struct {
	Language string `json:"language"` // "ru" or "kk"
}

// GenerateReply generates AI response for a review
// POST /api/v1/reviews/:id/generate-reply
func (h *ReviewHandler) GenerateReply(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	reviewID := c.Param("id")

	var req GenerateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Language = "ru" // Default to Russian
	}

	// Get review
	review, err := h.reviewRepo.GetByID(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Verify ownership
	if review.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Override language if specified
	if req.Language != "" {
		review.Language = req.Language
	}

	// Generate AI response
	aiResponse, err := h.aiResponder.GenerateResponse(review)
	if err != nil {
		logger.Log.Error("Failed to generate AI response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI response"})
		return
	}

	// Update review with AI response
	review.AIResponse = aiResponse
	if err := h.reviewRepo.Update(review); err != nil {
		logger.Log.Error("Failed to save AI response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save AI response"})
		return
	}

	// Get updated review
	review, _ = h.reviewRepo.GetByID(reviewID)

	c.JSON(http.StatusOK, gin.H{
		"message":     "AI response generated successfully",
		"review":      review,
		"ai_response": aiResponse,
	})
}

// UpdateReplyRequest represents request to update AI reply
type UpdateReplyRequest struct {
	AIResponse string `json:"ai_response" binding:"required"`
}

// UpdateReply updates AI response for a review (manual edit)
// PATCH /api/v1/reviews/:id/reply
func (h *ReviewHandler) UpdateReply(c *gin.Context) {
	telegramID := middleware.GetUserID(c)
	reviewID := c.Param("id")

	var req UpdateReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Get review
	review, err := h.reviewRepo.GetByID(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	// Verify ownership
	if review.UserID != telegramID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update AI response
	review.AIResponse = req.AIResponse
	if err := h.reviewRepo.Update(review); err != nil {
		logger.Log.Error("Failed to update AI response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update AI response"})
		return
	}

	// Get updated review
	review, _ = h.reviewRepo.GetByID(reviewID)

	c.JSON(http.StatusOK, gin.H{
		"message": "AI response updated successfully",
		"review":  review,
	})
}
