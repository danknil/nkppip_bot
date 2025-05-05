package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func scheduleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Debug(fmt.Sprintf("Opened schedule handler: %s", update.CallbackQuery.Data))
}

func buildScheduleKeyboard() models.ReplyMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{}},
	}

	return kb
}
