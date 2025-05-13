package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"os" "slices"
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

	var groups map[string]Group

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

			var currDay string
			var lessons []Lesson
			var days []Day

			var Index string
			var Time string
			var Name string
			var Teacher string
			var Cabinet string

			doc.
				Find("table.inf").
				First().
				Find("td.nul,td.hd0[colspan=4],td.hd,a.z1,a.z2,a.z3").
				Each(func(i int, s *goquery.Selection) {
					if s.HasClass("hd") {
						// День недели
						attr, exists := s.Attr("rowspan")
						if exists && attr == "6" {
							currDay = s.Text()
							slog.Debug(fmt.Sprintf("Текущий день: %s", currDay))
							return
						}
						// Номер пары
						if strings.Contains(s.Text(), "Пара") {
							lessons = append(lessons, Lesson{
								Index:   Index,
								Time:    Time,
								Name:    Name,
								Teacher: Teacher,
								Cabinet: Cabinet,
							})

							Index = ""
							Time = ""
							Name = ""
							Teacher = ""
							Cabinet = ""

							text := strings.Split(s.Text(), "<br>")
							slog.Debug(fmt.Sprintf("Получена пара: %+v", text))
							Index = strings.Split(text[0], " ")[0]
							Time = text[1]
						}
					}
					// Пустая строчка в паре
					if s.HasClass("nul") {
						Index = ""
						Time = ""
						slog.Debug("Сбрасываем пару, потому что пустое место")
						return
					}
					// Граница между днями
					if s.HasClass("hd0") {
						slog.Debug(fmt.Sprintf("Заканчиваем парсинг дня: %s", currDay))
						days = append(days, Day{
							DayOfWeek: currDay,
							Lessons:   lessons,
						})
						lessons = []Lesson{}
						return
					}
					if s.HasClass("z1") {
						Name = s.Text()
					}
					if s.HasClass("z2") {
						Cabinet = s.Text()
					}
					if s.HasClass("z3") {
						Teacher = s.Text()
					}
				})

			slog.Debug(fmt.Sprintf("Закончен парсинг группы: %s", group))

			Id := make([]byte, 16)
			for i := range Id {
				Id[i] = letters[rand.Int63()%int64(len(letters))]
			}
			groups[fmt.Sprintf("schedule_%s", string(Id))] = Group{
				Name: group,
				Days: days,
			}
		}
	})
	slog.Debug("Вывод групп:")
	for i, group := range groups {
		slog.Debug(fmt.Sprintf("%s Группа: %+v", i, group))
	}
	return groups
}

func scheduleHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if len(Groups) == 0 {
		return
	}
	slog.Debug(fmt.Sprintf("Opened schedule handler: %s", update.CallbackQuery.Data))

	if group, ok := Groups[update.CallbackQuery.Data]; ok {
		slog.Info(fmt.Sprintf("Найдена группа: %s", group.Name))

		messageBuilder := strings.Builder{}

		for _, day := range group.Days {
			messageBuilder.WriteString("\n**")
			messageBuilder.WriteString(strings.ToUpper(day.DayOfWeek))
			messageBuilder.WriteString(":**\n")
			for _, lesson := range day.Lessons {
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
			}
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:      messageBuilder.String(),
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			ParseMode: models.ParseModeMarkdown,
		})
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
		InlineKeyboard: [][]models.InlineKeyboardButton{{}},
	}

	return kb
}
