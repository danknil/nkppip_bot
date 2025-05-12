package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Group struct {
	Name string
	Days []Day
}

type Day struct {
	DayOfWeek string
	Lessons   []Lesson
}

type Lesson struct {
	TimeFrom string
	TimeTo   string
	Name     string
	Teacher  string
}

func scheduleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Debug(fmt.Sprintf("Opened schedule handler: %s", update.CallbackQuery.Data))

	// TODO: read from correct file
	body, err := os.ReadFile("")
	if err != nil {
		slog.Error("Не удалось прочитать расписание")
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		slog.Error("Не удалось получить документ расписания из файла")
		return
	}

	var groups []Group

	doc.Find("table.inf").First().Find("a.z0").Each(func(i int, sel *goquery.Selection) {
		groupRef, exists := sel.Attr("href")
		if exists {
			group := sel.Text()
			slog.Debug(fmt.Sprintf("Получено расписание %s: %s", group, groupRef))

			// TODO: read from correct file
			body, err := os.ReadFile(fmt.Sprintf("%s", groupRef))
			if err != nil {
				slog.Error(fmt.Sprintf("Не найдено расписание группы %s", group))
				return
			}
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				slog.Error(fmt.Sprintf("Не удалось получить документ расписания группы %s", group))
				return
			}

			var lessons []Lesson
			var days []Day

			doc.
				Find("table.inf").
				First().
				Find("td.ur[onmouseover][onmouseout],td.nul,td.hd0[colspan=4]").
				Each(func(i int, s *goquery.Selection) {
					if s.HasClass("nul") {
						return
					}
					if s.HasClass("hd0") {
						days = append(days, Day{
							// TODO: день недели
							DayOfWeek: "Пн",
							Lessons:   lessons,
						})
						lessons = []Lesson{}
						return
					}
					// TODO: обработка пар
				})
		}
	})
}

func buildScheduleKeyboard() models.ReplyMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{}},
	}

	return kb
}
