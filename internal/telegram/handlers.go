package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	user, err := b.getOrCreateUser(message.From)
	if err != nil {
		logger.Log.Error("Failed to get or create user", zap.Error(err))
		return
	}

	// Check if user is in a conversation state
	state := b.getUserState(message.Chat.ID)
	if state.AwaitingResponse {
		b.handleStateResponse(message, user, state)
		return
	}

	// Handle commands and menu buttons
	switch message.Text {
	case "/start":
		b.handleStart(message, user)
	case "📊 Dashboard":
		b.handleDashboard(message, user)
	case "📦 Low Stock Alerts":
		b.handleLowStockAlerts(message, user)
	case "⭐ Reviews":
		b.handleReviews(message, user)
	// TEMPORARILY DISABLED
	/*
	case "💰 Price Dumping":
		b.handlePriceDumping(message, user)
	*/
	case "🔑 Manage API Keys":
		b.handleManageAPIKeys(message, user)
	case "⚙️ Settings":
		b.handleSettings(message, user)
	case "ℹ️ Help":
		b.handleHelp(message, user)
	case "❌ Cancel":
		b.clearUserState(message.Chat.ID)
		b.sendMessageWithKeyboard(message.Chat.ID, "Cancelled.", GetMainMenuKeyboard())
	default:
		b.sendMessage(message.Chat.ID, "Please use the menu buttons or /start to begin.")
	}
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Answer callback query immediately
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	user, err := b.getOrCreateUser(query.From)
	if err != nil {
		logger.Log.Error("Failed to get user", zap.Error(err))
		return
	}

	parts := strings.Split(query.Data, ":")

	switch parts[0] {
	case "add_kaspi_key":
		b.handleAddKaspiKey(query.Message.Chat.ID, user)
	case "view_kaspi_key":
		b.handleViewKaspiKey(query.Message.Chat.ID, user)
	case "delete_kaspi_key":
		b.handleDeleteKaspiKey(query.Message.Chat.ID, user)
	case "toggle_auto_reply":
		b.handleToggleAutoReply(query.Message.Chat.ID, user)
	// TEMPORARILY DISABLED
	/*
	case "toggle_auto_dumping":
		b.handleToggleAutoDumping(query.Message.Chat.ID, user)
	*/
	case "view_dumping_products":
		b.handleViewDumpingProducts(query.Message.Chat.ID, user)
	case "enable_dumping":
		b.handleEnableDumpingPrompt(query.Message.Chat.ID, user)
	case "disable_dumping":
		b.handleDisableDumpingPrompt(query.Message.Chat.ID, user)
	case "change_language":
		b.handleChangeLanguage(query.Message.Chat.ID, user)
	case "lang":
		if len(parts) > 1 {
			b.handleSetLanguage(query.Message.Chat.ID, user, parts[1])
		}
	case "back_to_menu":
		b.handleBackToMenu(query.Message.Chat.ID)
	case "back_to_settings":
		b.handleSettings(query.Message, user)
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message, user *domain.User) {
	welcomeText := fmt.Sprintf(`Welcome to *Kaspi Seller Assistant*! 👋

Hello %s! I'm your personal assistant for managing your Kaspi.kz inventory and reviews.

*What I can do for you:*
📊 Track your inventory and predict days of stock
📦 Alert you when products are running low
⭐ Manage customer reviews with AI-powered responses
🤖 Auto-respond to reviews (if enabled)

*Getting Started:*
1. Add your Kaspi API key (🔑 Manage API Keys)
2. I'll automatically sync your products and sales data
3. Check your dashboard to see insights

Use the menu below to get started!`, user.FirstName)

	b.sendMessageWithKeyboard(message.Chat.ID, welcomeText, GetMainMenuKeyboard())
}

func (b *Bot) handleDashboard(message *tgbotapi.Message, user *domain.User) {
	// Get low stock products
	lowStockProducts, err := b.inventoryService.GetLowStockSummary(user.TelegramID, 7)
	if err != nil {
		logger.Log.Error("Failed to get low stock products", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to load dashboard. Please try again.")
		return
	}

	// Get pending reviews
	pendingReviews, err := b.reviewRepo.GetPendingReviews(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get pending reviews", zap.Error(err))
	}

	// Get all products count
	allProducts, err := b.productRepo.GetByUserID(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
	}

	dashboardText := fmt.Sprintf(`📊 *Dashboard*

*Overview:*
📦 Total Products: %d
⚠️ Low Stock Alerts: %d
⭐ Pending Reviews: %d
🤖 Auto-Reply: %s

*Quick Stats:*`,
		len(allProducts),
		len(lowStockProducts),
		len(pendingReviews),
		map[bool]string{true: "✅ Enabled", false: "❌ Disabled"}[user.AutoReplyEnabled],
	)

	if len(lowStockProducts) > 0 {
		dashboardText += "\n\n*Top 3 Low Stock Items:*\n"
		for i, product := range lowStockProducts {
			if i >= 3 {
				break
			}
			dashboardText += fmt.Sprintf("\n%d. *%s*\n   Stock: %d units | Days left: %d\n",
				i+1, product.Name, product.CurrentStock, product.DaysOfStock)
		}
	}

	if len(pendingReviews) > 0 {
		dashboardText += fmt.Sprintf("\n\n💡 You have %d reviews waiting for responses!", len(pendingReviews))
	}

	b.sendMessage(message.Chat.ID, dashboardText)
}

func (b *Bot) handleLowStockAlerts(message *tgbotapi.Message, user *domain.User) {
	products, err := b.inventoryService.GetLowStockSummary(user.TelegramID, 7)
	if err != nil {
		logger.Log.Error("Failed to get low stock products", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to load low stock alerts. Please try again.")
		return
	}

	if len(products) == 0 {
		b.sendMessage(message.Chat.ID, "✅ Great! No low stock alerts at the moment.\n\nAll your products have sufficient inventory.")
		return
	}

	alertText := fmt.Sprintf("📦 *Low Stock Alerts* (≤7 days)\n\nYou have %d product(s) running low:\n\n", len(products))

	for i, product := range products {
		emoji := "🟡"
		if product.DaysOfStock <= 3 {
			emoji = "🔴"
		}

		alertText += fmt.Sprintf("%s *%s*\n", emoji, product.Name)
		alertText += fmt.Sprintf("   • Current Stock: %d units\n", product.CurrentStock)
		alertText += fmt.Sprintf("   • Sales Velocity: %.1f units/day\n", product.SalesVelocity)
		alertText += fmt.Sprintf("   • Days of Stock: %d days\n", product.DaysOfStock)
		alertText += fmt.Sprintf("   • SKU: %s\n\n", product.SKU)

		if i >= 9 { // Limit to 10 products to avoid message length issues
			alertText += fmt.Sprintf("...and %d more\n", len(products)-10)
			break
		}
	}

	b.sendMessage(message.Chat.ID, alertText)
}

func (b *Bot) handleReviews(message *tgbotapi.Message, user *domain.User) {
	reviews, err := b.reviewRepo.GetByUserID(user.TelegramID, 10)
	if err != nil {
		logger.Log.Error("Failed to get reviews", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to load reviews. Please try again.")
		return
	}

	if len(reviews) == 0 {
		b.sendMessage(message.Chat.ID, "You don't have any reviews yet.")
		return
	}

	reviewText := "⭐ *Recent Reviews*\n\n"

	for _, review := range reviews {
		stars := strings.Repeat("⭐", review.Rating)
		statusEmoji := "⏳"
		if review.AIResponseSent {
			statusEmoji = "✅"
		} else if review.AIResponse != "" {
			statusEmoji = "📝"
		}

		reviewText += fmt.Sprintf("%s %s *%s*\n", statusEmoji, stars, review.AuthorName)
		reviewText += fmt.Sprintf("_%s_\n", truncateString(review.Comment, 100))

		if review.AIResponse != "" {
			reviewText += fmt.Sprintf("\n💬 Response: _%s_\n", truncateString(review.AIResponse, 100))
		}

		reviewText += "\n---\n\n"
	}

	pendingCount := 0
	for _, r := range reviews {
		if !r.AIResponseSent {
			pendingCount++
		}
	}

	if pendingCount > 0 {
		reviewText += fmt.Sprintf("💡 You have %d pending reviews.", pendingCount)
	}

	b.sendMessage(message.Chat.ID, reviewText)
}

func (b *Bot) handleManageAPIKeys(message *tgbotapi.Message, user *domain.User) {
	text := `🔑 *Manage Kaspi API Key*

Add your Kaspi.kz API key to start syncing your inventory and reviews.

Select an action below:`

	b.sendMessageWithKeyboard(message.Chat.ID, text, GetKaspiKeyboard())
}

func (b *Bot) handleAddKaspiKey(chatID int64, user *domain.User) {
	state := &UserState{
		State:            "adding_kaspi_key",
		Data:             make(map[string]interface{}),
		AwaitingResponse: true,
	}
	b.setUserState(chatID, state)

	text := `Adding *Kaspi.kz* API Key

Please send your API credentials in the following format:

API_KEY MERCHANT_ID

Example:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... 12345678

Send "Cancel" to abort.`

	b.sendMessageWithKeyboard(chatID, text, GetCancelKeyboard())
}

func (b *Bot) handleViewKaspiKey(chatID int64, user *domain.User) {
	key, err := b.kaspiKeyRepo.GetByUserID(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get Kaspi key", zap.Error(err))
		b.sendMessage(chatID, "Failed to load API key.")
		return
	}

	if key == nil {
		b.sendMessage(chatID, "You don't have a Kaspi API key configured yet.\n\nUse 'Add Kaspi Key' to get started!")
		return
	}

	status := "✅ Active"
	if !key.IsActive {
		status = "❌ Inactive"
	}

	text := fmt.Sprintf(`🔑 *Your Kaspi API Key*

Status: %s
Merchant ID: %s
Added: %s`, status, key.MerchantID, key.CreatedAt.Format("2006-01-02 15:04"))

	b.sendMessage(chatID, text)
}

func (b *Bot) handleDeleteKaspiKey(chatID int64, user *domain.User) {
	err := b.kaspiKeyRepo.Delete(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to delete Kaspi key", zap.Error(err))
		b.sendMessage(chatID, "Failed to delete API key.")
		return
	}

	b.sendMessage(chatID, "✅ Kaspi API key deleted successfully!")
}

func (b *Bot) handleSettings(message *tgbotapi.Message, user *domain.User) {
	text := `⚙️ *Settings*

Configure your bot preferences below:`

	b.sendMessageWithKeyboard(message.Chat.ID, text, GetSettingsKeyboard(user.AutoReplyEnabled, user.AutoDumpingEnabled))
}

func (b *Bot) handleToggleAutoReply(chatID int64, user *domain.User) {
	newState := !user.AutoReplyEnabled

	if err := b.userRepo.ToggleAutoReply(user.TelegramID, newState); err != nil {
		logger.Log.Error("Failed to toggle auto-reply", zap.Error(err))
		b.sendMessage(chatID, "Failed to update settings.")
		return
	}

	user.AutoReplyEnabled = newState

	status := "enabled"
	if !newState {
		status = "disabled"
	}

	text := fmt.Sprintf("✅ Auto-reply %s!", status)
	b.sendMessageWithKeyboard(chatID, text, GetSettingsKeyboard(user.AutoReplyEnabled, user.AutoDumpingEnabled))
}

func (b *Bot) handleChangeLanguage(chatID int64, user *domain.User) {
	text := "🌐 *Choose Language*\n\nSelect your preferred language for AI responses:"
	b.sendMessageWithKeyboard(chatID, text, GetLanguageKeyboard())
}

func (b *Bot) handleSetLanguage(chatID int64, user *domain.User, lang string) {
	user.LanguageCode = lang

	if err := b.userRepo.Update(user); err != nil {
		logger.Log.Error("Failed to update language", zap.Error(err))
		b.sendMessage(chatID, "Failed to update language.")
		return
	}

	langName := map[string]string{
		"ru": "Русский",
		"kk": "Қазақша",
		"en": "English",
	}[lang]

	b.sendMessage(chatID, fmt.Sprintf("✅ Language changed to %s", langName))
}

func (b *Bot) handleHelp(message *tgbotapi.Message, user *domain.User) {
	helpText := `ℹ️ *Help & Support*

*How to use this bot:*

1️⃣ *Add API Keys*
   Go to "🔑 Manage API Keys" and add your marketplace credentials.

2️⃣ *Sync Data*
   The bot automatically syncs your products, sales, and reviews every 6 hours.

3️⃣ *Monitor Inventory*
   Check "📦 Low Stock Alerts" to see products running low.

4️⃣ *Manage Reviews*
   View and respond to customer reviews with AI assistance.

5️⃣ *Enable Auto-Reply*
   Go to "⚙️ Settings" to enable automatic AI responses to reviews.

*Questions or Issues?*
Contact support: @your_support_username`

	b.sendMessage(message.Chat.ID, helpText)
}

func (b *Bot) handleBackToMenu(chatID int64) {
	b.sendMessageWithKeyboard(chatID, "Main Menu", GetMainMenuKeyboard())
}

func (b *Bot) handleStateResponse(message *tgbotapi.Message, user *domain.User, state *UserState) {
	if message.Text == "❌ Cancel" {
		b.clearUserState(message.Chat.ID)
		b.sendMessageWithKeyboard(message.Chat.ID, "Cancelled.", GetMainMenuKeyboard())
		return
	}

	switch state.State {
	case "adding_kaspi_key":
		b.processAddKaspiKey(message, user, state)
	case "enabling_dumping":
		b.processEnableDumping(message, user, state)
	case "disabling_dumping":
		b.processDisableDumping(message, user, state)
	}
}

func (b *Bot) processAddKaspiKey(message *tgbotapi.Message, user *domain.User, state *UserState) {
	parts := strings.Fields(message.Text)

	if len(parts) < 2 {
		b.sendMessage(message.Chat.ID, "Invalid format. Please provide both API_KEY and MERCHANT_ID.\n\nFormat: API_KEY MERCHANT_ID")
		return
	}

	apiKey := parts[0]
	merchantID := parts[1]

	// Encrypt API key
	encryptedKey, err := b.encryptor.Encrypt(apiKey)
	if err != nil {
		logger.Log.Error("Failed to encrypt API key", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to save API key. Please try again.")
		return
	}

	// Create Kaspi key
	key := &domain.KaspiKey{
		UserID:          user.TelegramID,
		APIKeyEncrypted: encryptedKey,
		MerchantID:      merchantID,
		IsActive:        true,
	}

	if err := b.kaspiKeyRepo.Create(key); err != nil {
		logger.Log.Error("Failed to create Kaspi key", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to save API key. This user may already have a key configured.")
		return
	}

	b.clearUserState(message.Chat.ID)

	// Trigger immediate sync
	go func() {
		if err := b.syncService.SyncUserData(key); err != nil {
			logger.Log.Error("Failed to sync Kaspi data", zap.Error(err))
		}
	}()

	text := "✅ Kaspi API key added successfully!\n\nYour data is now being synced. This may take a few minutes."
	b.sendMessageWithKeyboard(message.Chat.ID, text, GetMainMenuKeyboard())
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
