package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"os"
	"slices"
	"strings"

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
	Index   string
	Time    string
	Name    string
	Teacher string
	Cabinet string
}

// for rand generation
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const SchedulePath = "./schedule"

var Groups map[string]Group

func getSchedule(schedulePath string) map[string]Group {
	body, err := os.ReadFile(fmt.Sprintf("%s/bg.htm", schedulePath))
	if err != nil {
		slog.Error("Не удалось прочитать расписание")
		return map[string]Group{}
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		slog.Error("Не удалось получить документ расписания из файла")
		return map[string]Group{}
	}

	groups := make(map[string]Group)

	doc.Find("table.inf").First().Find("a.z0").Each(func(i int, sel *goquery.Selection) {
		groupRef, exists := sel.Attr("href")
		if exists {
			group := sel.Text()
			slog.Debug(fmt.Sprintf("Получено расписание %s: %s", group, groupRef))

			// TODO: read from correct file
			body, err := os.ReadFile(fmt.Sprintf("%s/%s", schedulePath, groupRef))
			if err != nil {
				slog.Error(fmt.Sprintf("Не найдено расписание группы %s", group))
				return
			}
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				slog.Error(fmt.Sprintf("Не удалось получить документ расписания группы %s", group))
				return
			}

			var days []Day
			var lessons []Lesson
			doc.
				Find("table.inf").
				First().
				Find("TD.hd[rowspan='6']").
				Each(func(i int, s *goquery.Selection) {
					day := s.Text()

					row := s.Parent()
					for {
						lessonHeader := row.ChildrenFiltered("TD.hd:not([rowspan])")
						lessonInfo := strings.Split(lessonHeader.Text(), ":")

						if len(lessonInfo) > 1 {
							Index := strings.Split(lessonInfo[0], " ")[0]
							Time := lessonInfo[1]

							row.ChildrenFiltered("TD.ur").Each(func(i int, s *goquery.Selection) {
								Name := row.ChildrenFiltered("a.z1").Text()
								Cabinet := s.ChildrenFiltered("a.z2").Text()
								Teacher := s.ChildrenFiltered("a.z3").Text()
								lesson := Lesson{
									Index:   Index,
									Time:    Time,
									Name:    Name,
									Cabinet: Cabinet,
									Teacher: Teacher,
								}

								// slog.Debug(fmt.Sprintf("Lesson: %+v", lesson))
								lessons = append(lessons, lesson)
							})
						}
						row = row.Next()
						if _, ex := row.Children().Attr("colspan"); ex || row.Length() == 0 {
							break
						}
					}
					days = append(days, Day{
						DayOfWeek: day,
						Lessons:   lessons,
					})
					lessons = []Lesson{}
				})

			slog.Debug(fmt.Sprintf("Закончен парсинг группы: %s", group))

			Id := make([]byte, 16)
			for i := range Id {
				Id[i] = letters[rand.Int63()%int64(len(letters))]
			}

			slog.Debug(fmt.Sprintf("Сырой Id: %+v", Id))

			stringID := fmt.Sprintf("schedule_%s", string(Id))

			slog.Debug(fmt.Sprintf("ID: %s", string(Id)))
			groups[stringID] = Group{
				Name: group,
				Days: days,
			}
		}
	})
	return groups
}

func scheduleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Debug(fmt.Sprintf("Opened schedule handler: %s", update.CallbackQuery.Data))

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	if len(Groups) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			ReplyMarkup: buildCategoryKeyboard(),
		})
		return
	}

	if group, ok := Groups[update.CallbackQuery.Data]; ok {
		slog.Info(fmt.Sprintf("Найдена группа: %s", group.Name))

		messageBuilder := strings.Builder{}

		for _, day := range group.Days {
			messageBuilder = strings.Builder{}

			messageBuilder.WriteString("`")
			messageBuilder.WriteString(day.DayOfWeek)
			messageBuilder.WriteString("`")
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				Text:      messageBuilder.String(),
				ParseMode: models.ParseModeMarkdown,
			})

			for _, lesson := range day.Lessons {
				messageBuilder = strings.Builder{}
				messageBuilder.WriteString("`")
				messageBuilder.WriteString(lesson.Index)
				messageBuilder.WriteString(" | ")
				messageBuilder.WriteString(lesson.Time)
				messageBuilder.WriteString("` ")
				messageBuilder.WriteString(lesson.Name)
				messageBuilder.WriteString(", ")
				messageBuilder.WriteString(lesson.Cabinet)
				messageBuilder.WriteString(" кабинет. Учитель: ")
				messageBuilder.WriteString(lesson.Teacher)
				messageBuilder.WriteString("\n")
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
					Text:      messageBuilder.String(),
					ParseMode: models.ParseModeMarkdownV1,
				})
			}
		}
	}
}

func buildScheduleKeyboard() models.ReplyMarkup {
	Groups = getSchedule(SchedulePath)
	schedIds := slices.Collect(maps.Keys(Groups))

	rows := 3
	columns := (len(schedIds) / rows) + 1
	var inlineKeyboard [][]models.InlineKeyboardButton

	for i := range columns {
		var column []models.InlineKeyboardButton
		for j := range rows {
			index := (i+1)*(j+1) - 1

			if len(schedIds) <= index {
				break
			}

			schedId := schedIds[index]
			group := Groups[schedId]
			slog.Debug(fmt.Sprintf("Добавлен документ: %+v", group))
			column = append(column, models.InlineKeyboardButton{
				Text:         group.Name,
				CallbackData: schedId,
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

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}

	return kb
}
