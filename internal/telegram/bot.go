package telegram

import (
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/internal/service"
	"github.com/yourusername/seller-assistant/pkg/crypto"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

type Bot struct {
	api                 *tgbotapi.BotAPI
	userRepo            domain.UserRepository
	kaspiKeyRepo        domain.KaspiKeyRepository
	productRepo         domain.ProductRepository
	reviewRepo          domain.ReviewRepository
	inventoryService    *service.InventoryService
	aiResponder         *service.AIResponderService
	syncService         *service.KaspiSyncService
	priceDumpingService *service.PriceDumpingService
	encryptor           *crypto.Encryptor

	// User state management for multi-step conversations
	userStates map[int64]*UserState
	stateMutex sync.RWMutex
}

type UserState struct {
	State            string
	Data             map[string]interface{}
	CurrentCommand   string
	AwaitingResponse bool
}

func NewBot(
	token string,
	userRepo domain.UserRepository,
	kaspiKeyRepo domain.KaspiKeyRepository,
	productRepo domain.ProductRepository,
	reviewRepo domain.ReviewRepository,
	inventoryService *service.InventoryService,
	aiResponder *service.AIResponderService,
	syncService *service.KaspiSyncService,
	priceDumpingService *service.PriceDumpingService,
	encryptor *crypto.Encryptor,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	logger.Log.Info("Telegram bot authorized", zap.String("username", api.Self.UserName))

	return &Bot{
		api:                 api,
		userRepo:            userRepo,
		kaspiKeyRepo:        kaspiKeyRepo,
		productRepo:         productRepo,
		reviewRepo:          reviewRepo,
		inventoryService:    inventoryService,
		aiResponder:         aiResponder,
		syncService:         syncService,
		priceDumpingService: priceDumpingService,
		encryptor:           encryptor,
		userStates:          make(map[int64]*UserState),
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	logger.Log.Info("Bot started, listening for updates...")

	for update := range updates {
		if update.Message != nil {
			go b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}

func (b *Bot) getUserState(chatID int64) *UserState {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()

	if state, ok := b.userStates[chatID]; ok {
		return state
	}

	return &UserState{
		State: "idle",
		Data:  make(map[string]interface{}),
	}
}

func (b *Bot) setUserState(chatID int64, state *UserState) {
	b.stateMutex.Lock()
	defer b.stateMutex.Unlock()

	b.userStates[chatID] = state
}

func (b *Bot) clearUserState(chatID int64) {
	b.stateMutex.Lock()
	defer b.stateMutex.Unlock()

	delete(b.userStates, chatID)
}

func (b *Bot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	switch k := keyboard.(type) {
	case tgbotapi.ReplyKeyboardMarkup:
		msg.ReplyMarkup = k
	case tgbotapi.InlineKeyboardMarkup:
		msg.ReplyMarkup = k
	}

	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) getOrCreateUser(from *tgbotapi.User) (*domain.User, error) {
	user, err := b.userRepo.GetByTelegramID(from.ID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = &domain.User{
			TelegramID:       from.ID,
			Username:         from.UserName,
			FirstName:        from.FirstName,
			LastName:         from.LastName,
			LanguageCode:     from.LanguageCode,
			AutoReplyEnabled: false,
		}

		if err := b.userRepo.Create(user); err != nil {
			return nil, err
		}

		logger.Log.Info("New user created",
			zap.Int64("telegram_id", from.ID),
			zap.String("username", from.UserName),
		)
	}

	return user, nil
}
