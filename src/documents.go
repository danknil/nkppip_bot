package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gopkg.in/yaml.v3"
)

type DocumentList struct {
	docs []Document `yaml:"documents"`
}

type Document struct {
	Id   string `yaml:"id"`
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func getDocuments(path string) []Document {
	var doclist []Document

	// return empty list on error
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Не удалось прочитать список документов")
		return []Document{}
	}
	slog.Debug(string(yamlFile))
	err = yaml.Unmarshal(yamlFile, &doclist)
	if err != nil {
		slog.Error("Не удалось получить документы")
	}

	slog.Debug(fmt.Sprintf("Документов получено: %d", len(doclist)))
	for _, doc := range doclist {
		slog.Debug(fmt.Sprintf("Получен документ: %+v", doc))
	}

	return doclist
}

func documentHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Debug(fmt.Sprintf("Opened document handler: %s", update.CallbackQuery.Data))

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	var doc Document

	doc_id := update.CallbackQuery.Data
	docs := getDocuments("./docs/list.yml")

	for i, pdoc := range docs {
		pdoc_id := fmt.Sprintf("document_%s", pdoc.Id)
		slog.Debug(fmt.Sprintf("pdoc_id: %s", pdoc_id))
		if pdoc_id == doc_id {
			doc = docs[i]
			slog.Info(fmt.Sprintf("Выбран документ: %s", doc.Name))
			break
		}
	}

	fileData, err := os.ReadFile(fmt.Sprintf("./docs/%s", doc.Path))
	if err != nil {
		slog.Error("Не удалось прочитать документ")
		return
	}

	filename := filepath.Base(doc.Path)
	slog.Info(fmt.Sprintf("Отправлен документ: %s", filename)
	b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Document: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(fileData),
		},
		Caption: doc.Name,
	})
}

func buildDocumentKeyboard() models.ReplyMarkup {
	var inlineKeyboard [][]models.InlineKeyboardButton

	docs := getDocuments("./docs/list.yml")
	rows := 3
	columns := (len(docs) / rows) + 1

	for i := range columns {
		var column []models.InlineKeyboardButton
		for j := range rows {
			index := (i+1)*(j+1) - 1

			if len(docs) <= index {
				break
			}

			doc := docs[index]
			slog.Debug(fmt.Sprintf("Добавлен документ: %+v", doc))
			column = append(column, models.InlineKeyboardButton{
				Text:         doc.Name,
				CallbackData: fmt.Sprintf("document_%s", doc.Id),
			})
		}
		if len(column) != 0 {
			slog.Debug(fmt.Sprintf("Длина колонки: %d", len(column)))
			inlineKeyboard = append(inlineKeyboard, column)
		}
	}

	inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{
		Text:         "<<",
		CallbackData: "prev_btn",
	}})

	slog.Debug(fmt.Sprintf("Список документов: %+v", inlineKeyboard))

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	return kb
}
