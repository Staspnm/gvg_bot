package handlers

import (
	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

func InitHandlers(bot *telebot.Bot, db *database.Database) {
	// Регистрация пользователя
	bot.Handle("/register", func(c telebot.Context) error {
		return handleRegistration(c, db)
	})

	// Деактивация пользователя
	bot.Handle("/deactivate", func(c telebot.Context) error {
		return handleDeactivateUser(c, db)
	})

	// Отчет о битве (обрабатываем все сообщения, не начинающиеся с /)
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Message().Text[0] != '/' {
			return handleBattleReport(c, db)
		}
		return nil
	})

	// Команды для просмотра результатов по локациям
	locations := []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
	for _, loc := range locations {
		bot.Handle("/"+loc, func(c telebot.Context) error {
			return handleBattleResults(c, db)
		})
	}
}
