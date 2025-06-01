package handlers

import (
	"fmt"
	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
	"strings"
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
	// Добавляем функцию смены роли
	bot.Handle("/setrole", func(c telebot.Context) error {
		return handleSetRole(c, db)
	})

	// Добавляем в функцию InitHandlers
	bot.Handle("/editmyinfo", func(c telebot.Context) error {
		return handleEditUserSelf(c, db)
	})

	bot.Handle("/edituser", func(c telebot.Context) error {
		return handleEditUserByOfficer(c, db)
	})

	// Добавляем в функцию InitHandlers
	bot.Handle("/missingreports", func(c telebot.Context) error {
		return handleMissingReports(c, db, "")
	})

	// Добавляем команды для каждой локации
	locations := []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
	for _, loc := range locations {
		bot.Handle("/missingreports"+loc, func(c telebot.Context) error {
			loc := strings.TrimPrefix(c.Message().Text, "/missingreports")
			return handleMissingReports(c, db, loc)
		})
	}

	bot.Handle("/changeguild", func(c telebot.Context) error {
		return handleChangeGuild(c, db)
	})

	// Добавляем обработчик для кнопки
	bot.Handle(&telebot.Btn{Text: "Напомнить игрокам"}, func(c telebot.Context) error {
		missingPlayers, ok := c.Get("missing_players").([]string)
		if !ok || len(missingPlayers) == 0 {
			return c.Send("Не удалось найти игроков для напоминания.")
		}

		// Формируем сообщение с упоминаниями
		var mentions []string
		for _, player := range missingPlayers {
			// Находим Telegram ID игрока по никнейму
			var tgID int64
			err := db.QueryRow(`
            SELECT telegram_id FROM users 
            WHERE game_nickname = $1 AND guild_name = $2
        `, player, c.Sender().Username).Scan(&tgID)

			if err == nil && tgID != 0 {
				mentions = append(mentions, fmt.Sprintf("@%d", tgID))
			} else {
				mentions = append(mentions, player)
			}
		}

		msg := fmt.Sprintf("⏰ Напоминание: пожалуйста, отправьте отчеты!\n\nНе отчитались:\n%s",
			strings.Join(mentions, "\n"))

		return c.Send(msg)
	})

	// Отчет о битве (обрабатываем все сообщения, не начинающиеся с /)
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Message().Text[0] != '/' {
			return handleBattleReport(c, db)
		}
		return nil
	})

	// Команды для просмотра результатов по локациям
	locations = []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
	for _, loc := range locations {
		bot.Handle("/"+loc, func(c telebot.Context) error {
			return handleBattleResults(c, db)
		})
	}

	// Добавляем в функцию InitHandlers
	bot.Handle("/userinfo", func(c telebot.Context) error {
		return handleUserInfo(c, db)
	})

}
