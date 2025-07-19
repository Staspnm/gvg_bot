package handlers

import (
	"fmt"
	"strings"

	"github.com/gvg-bot/database"
	"github.com/gvg-bot/usecases/registration"
	"gopkg.in/telebot.v3"
)

type registrator interface {
	Registration(user registration.User, db *database.Database) error
}

type reporter interface {
	MissingReports(db *database.Database) ([]string, error)
}

type Handler struct {
	registrator registrator
	reporter    reporter
	bot         *telebot.Bot
	db          *database.Database
}

func NewHandlers(bot *telebot.Bot, db *database.Database, registrator registrator, reporter reporter) *Handler {
	return &Handler{registrator: registrator, reporter: reporter, bot: bot, db: db}
}

func (handler *Handler) InitHandlers() {
	// Регистрация пользователя
	handler.bot.Handle("/register", func(c telebot.Context) error {
		return handler.handleRegistration(c, handler.db)
	})

	// Деактивация пользователя
	handler.bot.Handle("/deactivate", func(c telebot.Context) error {
		return handleDeactivateUser(c, handler.db)
	})
	// Добавляем функцию смены роли
	handler.bot.Handle("/setrole", func(c telebot.Context) error {
		return handler.handleSetRole(c, handler.db)
	})

	// Добавляем в функцию InitHandlers
	handler.bot.Handle("/editmyinfo", func(c telebot.Context) error {
		return handler.handleEditUserSelf(c, handler.db)
	})

	handler.bot.Handle("/edituser", func(c telebot.Context) error {
		return handler.handleEditUserByOfficer(c, handler.db)
	})

	// Добавляем в функцию InitHandlers
	handler.bot.Handle("/missingreports", func(c telebot.Context) error {
		return handler.handleMissingReports(c, handler.db, "")
	})

	// Добавляем команды для каждой локации
	locations := []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
	for _, loc := range locations {
		handler.bot.Handle("/missingreports"+loc, func(c telebot.Context) error {
			loc := strings.TrimPrefix(c.Message().Text, "/missingreports")
			return handler.handleMissingReports(c, handler.db, loc)
		})
	}

	handler.bot.Handle("/changeguild", func(c telebot.Context) error {
		return handler.handleChangeGuild(c, handler.db)
	})

	// Добавляем обработчик для кнопки
	handler.bot.Handle(&telebot.Btn{Text: "Напомнить игрокам"}, func(c telebot.Context) error {
		missingPlayers, ok := c.Get("missing_players").([]string)

		if !ok || len(missingPlayers) == 0 {
			return c.Send("Не удалось найти игроков для напоминания.")
		}

		// Формируем сообщение с упоминаниями
		var mentions []string
		for _, player := range missingPlayers {
			// Находим Telegram ID игрока по никнейму
			var tgID int64
			err := handler.db.QueryRow(`
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
	handler.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		if c.Message().Text[0] != '/' {
			return handler.handleBattleReport(c, handler.db)
		}
		return nil
	})

	// Команды для просмотра результатов по локациям
	locations = []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
	for _, loc := range locations {
		handler.bot.Handle("/"+loc, func(c telebot.Context) error {
			return handler.handleBattleResults(c, handler.db)
		})
	}

	// Добавляем в функцию InitHandlers
	handler.bot.Handle("/userinfo", func(c telebot.Context) error {
		return handler.handleUserInfo(c, handler.db)
	})

}
