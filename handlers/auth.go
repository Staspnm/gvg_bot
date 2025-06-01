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

type BotHandler struct {
	registrator registrator
}

func New(registrator registrator) *BotHandler {
	return &BotHandler{
		registrator: registrator,
	}
}

func (handler *BotHandler) handleRegistration(c telebot.Context, db *database.Database) error {
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 5 {
		return c.Send("Неверный формат регистрации. Используйте: /register игровой_ник 123456789 название_гильдии роль")
	}

	gameNick := args[1]
	code := args[2]
	guildName := args[3]
	role := strings.ToLower(args[4])

	user := registration.User{
		TelegramID:    c.Sender().ID,
		GameNickname:  gameNick,
		NineDigitCode: code,
		GuildName:     guildName,
		GuildRole:     role,
	}

	err := handler.registrator.Registration(user, db)
	if err != nil {
		c.Send(fmt.Printf("Возникла ошибка регистрации: %v", err))
		return err
	}

	return c.Send(fmt.Sprintf("Регистрация успешна! Добро пожаловать, %s (%s) гильдии %s!", gameNick, role, guildName))
}

func handleDeactivateUser(c telebot.Context, db *database.Database) error {
	// Проверяем права (только owner/leader могут деактивировать)
	var requesterRole string
	err := db.QueryRow("SELECT guild_role FROM users WHERE telegram_id = $1", c.Sender().ID).Scan(&requesterRole)
	if err != nil {
		return c.Send("Вы не зарегистрированы.")
	}

	if requesterRole != "owner" && requesterRole != "leader" {
		return c.Send("У вас нет прав для этой команды.")
	}

	// Получаем ID пользователя для деактивации (из reply или упоминания)
	targetUser := c.Message().ReplyTo.Sender
	if targetUser == nil {
		return c.Send("Ответьте на сообщение пользователя или упомяните его.")
	}

	// Деактивируем пользователя
	_, err = db.Exec("UPDATE users SET is_active = FALSE WHERE telegram_id = $1", targetUser.ID)
	if err != nil {
		return c.Send("Ошибка при деактивации пользователя.")
	}

	return c.Send(fmt.Sprintf("Пользователь @%s деактивирован.", targetUser.Username))
}
