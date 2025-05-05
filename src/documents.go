package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gopkg.in/yaml.v3"
)

type Document struct {
	id   string `yaml:id`
	name string `yaml:"name"`
	path string `yaml:"path"`
}

func getDocuments(path string) []Document {
	var docs []Document

	// return empty list on error
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Can't read list file")
		return []Document{}
	}
	err = yaml.Unmarshal(yamlFile, &docs)
	if err != nil {
		slog.Error("Failed to get documents")
	}

	return docs
}

func documentHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("Open schedule handler: %s", update.CallbackQuery.Data))

}

func buildDocumentKeyboard() models.ReplyMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{}},
	}

	return kb
}
