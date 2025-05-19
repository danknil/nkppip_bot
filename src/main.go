package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type message struct {
	command     string   `yaml:"command"`
	description string   `yaml:"desc"`
	questions   []string `yaml:"questions"`
}

var botToken = os.Getenv("BOT_TOKEN")
var appEnv = os.Getenv("APP_ENV")

func main() {
	// setup logging
	log_opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	var handler slog.Handler = slog.NewTextHandler(os.Stdout, log_opts)
	if appEnv == "production" {
		log_opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, log_opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, startHandler),
		bot.WithCallbackQueryDataHandler("category_", bot.MatchTypePrefix, categoryHandler),
		bot.WithCallbackQueryDataHandler("prev_btn", bot.MatchTypeExact, prevHandler),
		bot.WithCallbackQueryDataHandler("document_", bot.MatchTypePrefix, documentHandler),
		bot.WithCallbackQueryDataHandler("schedule_", bot.MatchTypePrefix, scheduleHandler),
	}

	b, err := bot.New(botToken, opts...)
	if err != nil {
		slog.Error("Ошибка при создании объекта бота")
		panic(err)
	}

	slog.Info("Бот успешно запущен")

	b.Start(ctx)
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: buildCategoryKeyboard(),
	})
}

func prevHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: buildCategoryKeyboard(),
	})
}

func categoryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	var replyMarkup models.ReplyMarkup

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	slog.Debug(fmt.Sprintf("Got category: %s", update.CallbackQuery.Data))

	switch update.CallbackQuery.Data {
	case "category_schedule":
		replyMarkup = buildScheduleKeyboard()
	case "category_document":
		replyMarkup = buildDocumentKeyboard()
	default:
		replyMarkup = buildCategoryKeyboard()
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: replyMarkup,
	})
}

func buildCategoryKeyboard() models.ReplyMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Расписание", CallbackData: "category_schedule"},
				{Text: "Документы", CallbackData: "category_document"},
			},
		},
	}

	return kb
}
