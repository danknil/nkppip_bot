package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"

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
	slog.Debug(fmt.Sprintf("Opened document handler: %s", update.CallbackQuery.Data))

	var doc Document

	doc_id := update.CallbackQuery.Data
	docs := getDocuments("./docs/list.yaml")

	for i := 0; i < len(docs); i++ {
		if fmt.Sprintf("document_%s", docs[i].id) == doc_id {
			doc = docs[i]
			break
		}
	}

	slog.Info(fmt.Sprintf("Выбран документ: %s", doc.name))

	fileData, err := os.ReadFile(doc.path)
	if err != nil {
		slog.Error("Не удалось прочитать документ")
		return
	}

	filename := filepath.Base(doc.path)
	slog.Info(fmt.Sprintf("Отправлен документ: %s", filename))
	b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: update.CallbackQuery.Message.Message.ID,
		Document: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(fileData),
		},
		Caption: doc.name,
	})
}

func buildDocumentKeyboard() models.ReplyMarkup {
	var inlineKeyboard [][]models.InlineKeyboardButton

	docs := getDocuments("./docs/list.yaml")
	rows := math.Round(float64(len(docs)) / 8)

	for i := 0; i < int(rows); i++ {
		var column []models.InlineKeyboardButton
		for j := 0; j < (len(docs)/int(rows))+1; j++ {
			// check if index inbounds
			if len(docs) > i*j {
				doc := docs[i*j]
				column = append(column, models.InlineKeyboardButton{
					Text:         doc.name,
					CallbackData: fmt.Sprintf("document_%s", doc.id),
				})
			}
		}
		inlineKeyboard = append(inlineKeyboard, column)
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	return kb
}
