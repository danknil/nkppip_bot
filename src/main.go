package main

import (
	"context"
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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, startHandler),
		bot.WithCallbackQueryDataHandler("category_", bot.MatchTypePrefix, categoryHandler),
		// bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: buildCategoryKeyboard(),
	})
}

func categoryHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	var replyMarkup models.ReplyMarkup

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	switch update.CallbackQuery.Data {
	case "category_schedule":
		replyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "test1", CallbackData: "btn_opt1"},
					{Text: "test2", CallbackData: "btn_opt2"},
					{Text: "test3", CallbackData: "btn_opt3"},
				},
			},
		}
	case "category_documents":
		replyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "test1", CallbackData: "btn_opt1"},
					{Text: "test2", CallbackData: "btn_opt2"},
					{Text: "test3", CallbackData: "btn_opt3"},
				},
			},
		}
	default:
		replyMarkup = buildCategoryKeyboard()
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Выберите опцию",
		ReplyMarkup: replyMarkup,
	})
}

func buildCategoryKeyboard() models.ReplyMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Расписание", CallbackData: "category_schedule"},
				{Text: "Документы", CallbackData: "category_documents"},
			},
		},
	}

	return kb
}
