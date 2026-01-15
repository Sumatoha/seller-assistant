package telegram

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/seller-assistant/internal/domain"
	"github.com/yourusername/seller-assistant/pkg/logger"
	"go.uber.org/zap"
)

func (b *Bot) handlePriceDumping(message *tgbotapi.Message, user *domain.User) {
	text := `💰 *Price Dumping (Автодемпинг)*

Автоматически отслеживаем цены конкурентов и ставим цену на 1₸ дешевле каждые 5 минут.

*Как это работает:*
1️⃣ Включите автодемпинг для товара
2️⃣ Установите минимальную цену (защита от демпинга ниже себестоимости)
3️⃣ Система автоматически проверяет цены конкурентов
4️⃣ Если цена конкурента выше вашего минимума - обновляем на 1₸ дешевле

*Важно:* Если цена конкурентов падает ниже вашего минимума, система не будет обновлять цену.

Выберите действие:`

	b.sendMessageWithKeyboard(message.Chat.ID, text, GetPriceDumpingKeyboard())
}

func (b *Bot) handleToggleAutoDumping(chatID int64, user *domain.User) {
	newState := !user.AutoDumpingEnabled

	if err := b.userRepo.ToggleAutoDumping(user.TelegramID, newState); err != nil {
		logger.Log.Error("Failed to toggle auto-dumping", zap.Error(err))
		b.sendMessage(chatID, "Failed to update settings.")
		return
	}

	user.AutoDumpingEnabled = newState

	status := "включен"
	if !newState {
		status = "выключен"
	}

	text := fmt.Sprintf("✅ Автодемпинг %s!\n\n", status)
	if newState {
		text += "Теперь система будет автоматически отслеживать цены конкурентов каждые 5 минут.\n\n"
		text += "Не забудьте включить автодемпинг для конкретных товаров и установить минимальные цены!"
	} else {
		text += "Автоматическое обновление цен приостановлено."
	}

	b.sendMessageWithKeyboard(chatID, text, GetSettingsKeyboard(user.AutoReplyEnabled, user.AutoDumpingEnabled))
}

func (b *Bot) handleViewDumpingProducts(chatID int64, user *domain.User) {
	products, err := b.productRepo.GetByUserID(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		b.sendMessage(chatID, "Failed to load products.")
		return
	}

	if len(products) == 0 {
		b.sendMessage(chatID, "У вас пока нет товаров в системе.")
		return
	}

	// Фильтруем товары с включенным автодемпингом
	dumpingProducts := []domain.Product{}
	for _, p := range products {
		if p.AutoDumpingEnabled {
			dumpingProducts = append(dumpingProducts, p)
		}
	}

	text := "*Товары с автодемпингом:*\n\n"

	if len(dumpingProducts) == 0 {
		text += "Нет товаров с включенным автодемпингом.\n\n"
		text += "Используйте 'Enable for Product' чтобы включить автодемпинг."
	} else {
		for i, p := range dumpingProducts {
			if i >= 10 {
				text += fmt.Sprintf("\n...и еще %d товаров", len(dumpingProducts)-10)
				break
			}

			statusEmoji := "✅"
			if p.Price <= p.MinPrice {
				statusEmoji = "⚠️" // Цена на минимуме
			}

			text += fmt.Sprintf("%s *%s*\n", statusEmoji, p.Name)
			text += fmt.Sprintf("   • Текущая цена: %.0f₸\n", p.Price)
			text += fmt.Sprintf("   • Мин. цена: %.0f₸\n", p.MinPrice)
			if p.CompetitorMinPrice > 0 {
				text += fmt.Sprintf("   • Мин. цена конкурентов: %.0f₸\n", p.CompetitorMinPrice)
			}
			text += fmt.Sprintf("   • SKU: %s\n\n", p.SKU)
		}
	}

	// Также покажем несколько товаров без автодемпинга
	nonDumpingProducts := []domain.Product{}
	for _, p := range products {
		if !p.AutoDumpingEnabled {
			nonDumpingProducts = append(nonDumpingProducts, p)
		}
	}

	if len(nonDumpingProducts) > 0 {
		text += "\n*Другие товары (без автодемпинга):*\n\n"
		for i, p := range nonDumpingProducts {
			if i >= 5 {
				text += fmt.Sprintf("...и еще %d товаров\n", len(nonDumpingProducts)-5)
				break
			}
			text += fmt.Sprintf("• %s (SKU: %s)\n", p.Name, p.SKU)
		}
	}

	b.sendMessage(chatID, text)
}

func (b *Bot) handleEnableDumpingPrompt(chatID int64, user *domain.User) {
	state := &UserState{
		State:            "enabling_dumping",
		Data:             make(map[string]interface{}),
		AwaitingResponse: true,
	}
	b.setUserState(chatID, state)

	text := `Включить автодемпинг для товара

Отправьте данные в формате:
SKU МИНИМАЛЬНАЯ_ЦЕНА

Пример:
ABC123 15000

Где:
• SKU - артикул товара
• МИНИМАЛЬНАЯ_ЦЕНА - минимальная цена в тенге (ниже этой цены система не будет опускать цену)

Отправьте "Cancel" для отмены.`

	b.sendMessageWithKeyboard(chatID, text, GetCancelKeyboard())
}

func (b *Bot) handleDisableDumpingPrompt(chatID int64, user *domain.User) {
	state := &UserState{
		State:            "disabling_dumping",
		Data:             make(map[string]interface{}),
		AwaitingResponse: true,
	}
	b.setUserState(chatID, state)

	text := `Выключить автодемпинг для товара

Отправьте SKU товара (артикул).

Пример:
ABC123

Отправьте "Cancel" для отмены.`

	b.sendMessageWithKeyboard(chatID, text, GetCancelKeyboard())
}

func (b *Bot) processEnableDumping(message *tgbotapi.Message, user *domain.User, state *UserState) {
	parts := strings.Fields(message.Text)

	if len(parts) < 2 {
		b.sendMessage(message.Chat.ID, "Неверный формат. Отправьте: SKU МИНИМАЛЬНАЯ_ЦЕНА\n\nПример: ABC123 15000")
		return
	}

	sku := parts[0]
	minPriceStr := parts[1]

	minPrice, err := strconv.ParseFloat(minPriceStr, 64)
	if err != nil || minPrice <= 0 {
		b.sendMessage(message.Chat.ID, "Неверная минимальная цена. Укажите число больше 0.")
		return
	}

	// Найти товар по SKU
	products, err := b.productRepo.GetByUserID(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to find product.")
		return
	}

	var foundProduct *domain.Product
	for _, p := range products {
		if p.SKU == sku {
			foundProduct = &p
			break
		}
	}

	if foundProduct == nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Товар с SKU '%s' не найден.", sku))
		return
	}

	// Включить автодемпинг
	if err := b.priceDumpingService.EnableProductDumping(foundProduct.ID, minPrice); err != nil {
		logger.Log.Error("Failed to enable dumping", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to enable auto-dumping.")
		return
	}

	b.clearUserState(message.Chat.ID)

	text := fmt.Sprintf("✅ Автодемпинг включен!\n\n"+
		"*Товар:* %s\n"+
		"*SKU:* %s\n"+
		"*Минимальная цена:* %.0f₸\n\n"+
		"Система начнет отслеживать цены конкурентов каждые 5 минут.",
		foundProduct.Name, foundProduct.SKU, minPrice)

	b.sendMessageWithKeyboard(message.Chat.ID, text, GetMainMenuKeyboard())
}

func (b *Bot) processDisableDumping(message *tgbotapi.Message, user *domain.User, state *UserState) {
	sku := strings.TrimSpace(message.Text)

	// Найти товар по SKU
	products, err := b.productRepo.GetByUserID(user.TelegramID)
	if err != nil {
		logger.Log.Error("Failed to get products", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to find product.")
		return
	}

	var foundProduct *domain.Product
	for _, p := range products {
		if p.SKU == sku {
			foundProduct = &p
			break
		}
	}

	if foundProduct == nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Товар с SKU '%s' не найден.", sku))
		return
	}

	// Выключить автодемпинг
	if err := b.priceDumpingService.DisableProductDumping(foundProduct.ID); err != nil {
		logger.Log.Error("Failed to disable dumping", zap.Error(err))
		b.sendMessage(message.Chat.ID, "Failed to disable auto-dumping.")
		return
	}

	b.clearUserState(message.Chat.ID)

	text := fmt.Sprintf("✅ Автодемпинг выключен для товара:\n\n"+
		"*%s* (SKU: %s)\n\n"+
		"Цены больше не будут автоматически обновляться.",
		foundProduct.Name, foundProduct.SKU)

	b.sendMessageWithKeyboard(message.Chat.ID, text, GetMainMenuKeyboard())
}
