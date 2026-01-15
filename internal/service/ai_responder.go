package service

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type AIResponderService struct {
	openaiClient *openai.Client
	reviewRepo   domain.ReviewRepository
}

func NewAIResponderService(apiKey string, reviewRepo domain.ReviewRepository) *AIResponderService {
	return &AIResponderService{
		openaiClient: openai.NewClient(apiKey),
		reviewRepo:   reviewRepo,
	}
}

// GenerateResponse generates an AI response for a review
func (s *AIResponderService) GenerateResponse(review *domain.Review) (string, error) {
	prompt := s.buildPrompt(review)

	resp, err := s.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: s.getSystemPrompt(review.Language),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   300,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Choices[0].Message.Content, nil
}

// ProcessPendingReviews processes all pending reviews for a user
func (s *AIResponderService) ProcessPendingReviews(userID int64, autoSend bool) error {
	reviews, err := s.reviewRepo.GetPendingReviews(userID)
	if err != nil {
		return fmt.Errorf("failed to get pending reviews: %w", err)
	}

	logger.Log.Info("Processing pending reviews",
		zap.Int64("user_id", userID),
		zap.Int("count", len(reviews)),
	)

	for _, review := range reviews {
		response, err := s.GenerateResponse(&review)
		if err != nil {
			logger.Log.Error("Failed to generate AI response",
				zap.String("review_id", review.ID),
				zap.Error(err),
			)
			continue
		}

		review.AIResponse = response
		if autoSend {
			review.AIResponseSent = true
		}

		if err := s.reviewRepo.Update(&review); err != nil {
			logger.Log.Error("Failed to update review",
				zap.String("review_id", review.ID),
				zap.Error(err),
			)
			continue
		}

		logger.Log.Info("Generated AI response",
			zap.String("review_id", review.ID),
			zap.Bool("auto_sent", autoSend),
		)
	}

	return nil
}

// buildPrompt builds the prompt for AI response generation
func (s *AIResponderService) buildPrompt(review *domain.Review) string {
	ratingText := ""
	switch review.Rating {
	case 5:
		ratingText = "отличный отзыв (5 звезд)"
	case 4:
		ratingText = "хороший отзыв (4 звезды)"
	case 3:
		ratingText = "средний отзыв (3 звезды)"
	case 2:
		ratingText = "плохой отзыв (2 звезды)"
	case 1:
		ratingText = "очень плохой отзыв (1 звезда)"
	}

	prompt := fmt.Sprintf(
		"Покупатель %s оставил %s со следующим комментарием:\n\n\"%s\"\n\nНапишите профессиональный и дружелюбный ответ от имени продавца.",
		review.AuthorName,
		ratingText,
		review.Comment,
	)

	return prompt
}

// getSystemPrompt returns the system prompt based on language
func (s *AIResponderService) getSystemPrompt(language string) string {
	prompts := map[string]string{
		"ru": `Вы - профессиональный ассистент для продавцов на маркетплейсах. Ваша задача - генерировать вежливые, профессиональные и дружелюбные ответы на отзывы покупателей.

Правила:
1. Благодарите за отзыв
2. Если отзыв положительный - выражайте благодарность и радость
3. Если отзыв негативный - приносите извинения и предлагайте решение
4. Будьте кратки (2-4 предложения)
5. Используйте формальный, но дружелюбный тон
6. Не используйте эмодзи
7. Пишите на русском языке`,

		"kk": `Сіз маркетплейстердегі сатушыларға арналған кәсіби көмекшісіз. Сіздің міндетіңіз - сатып алушылардың пікірлеріне сыпайы, кәсіби және достық жауаптар жасау.

Ережелер:
1. Пікір үшін алғыс білдіріңіз
2. Егер пікір оң болса - алғыс пен қуанышты білдіріңіз
3. Егер пікір теріс болса - кешірім сұраңыз және шешім ұсыныңыз
4. Қысқа болыңыз (2-4 сөйлем)
5. Ресми, бірақ достық үнді пайдаланыңыз
6. Эмодзи қолданбаңыз
7. Қазақ тілінде жазыңыз`,

		"en": `You are a professional assistant for marketplace sellers. Your task is to generate polite, professional, and friendly responses to customer reviews.

Rules:
1. Thank them for the review
2. If positive - express gratitude and joy
3. If negative - apologize and offer a solution
4. Be brief (2-4 sentences)
5. Use a formal but friendly tone
6. Don't use emojis
7. Write in English`,
	}

	if prompt, ok := prompts[language]; ok {
		return prompt
	}

	return prompts["ru"] // Default to Russian
}
