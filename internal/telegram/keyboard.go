package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Dashboard"),
			tgbotapi.NewKeyboardButton("📦 Low Stock Alerts"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⭐ Reviews"),
			// tgbotapi.NewKeyboardButton("💰 Price Dumping"), // TEMPORARILY DISABLED
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔑 Manage API Keys"),
			tgbotapi.NewKeyboardButton("⚙️ Settings"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ℹ️ Help"),
		),
	)
}

func GetKaspiKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add Kaspi Key", "add_kaspi_key"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("View My Key", "view_kaspi_key"),
			tgbotapi.NewInlineKeyboardButtonData("Delete Key", "delete_kaspi_key"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "back_to_menu"),
		),
	)
}

func GetSettingsKeyboard(autoReplyEnabled bool, autoDumpingEnabled bool) tgbotapi.InlineKeyboardMarkup {
	autoReplyText := "Enable Auto-Reply"
	if autoReplyEnabled {
		autoReplyText = "Disable Auto-Reply"
	}

	// TEMPORARILY DISABLED - Auto Dumping
	/*
	autoDumpingText := "Enable Auto-Dumping"
	if autoDumpingEnabled {
		autoDumpingText = "Disable Auto-Dumping"
	}
	*/

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(autoReplyText, "toggle_auto_reply"),
		),
		// TEMPORARILY DISABLED
		/*
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(autoDumpingText, "toggle_auto_dumping"),
		),
		*/
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Change Language", "change_language"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "back_to_menu"),
		),
	)
}

func GetPriceDumpingKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("View Products", "view_dumping_products"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Enable for Product", "enable_dumping"),
			tgbotapi.NewInlineKeyboardButtonData("Disable for Product", "disable_dumping"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "back_to_menu"),
		),
	)
}

func GetLanguageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Русский", "lang:ru"),
			tgbotapi.NewInlineKeyboardButtonData("Қазақша", "lang:kk"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("English", "lang:en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back", "back_to_settings"),
		),
	)
}

func GetCancelKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Cancel"),
		),
	)
}
