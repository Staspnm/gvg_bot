package handlers

import (
	"fmt"
	"strings"

	"github.com/gvg-bot/database"
	"github.com/gvg-bot/models"
	"gopkg.in/telebot.v3"
)

func handleRegistration(c telebot.Context, db *database.Database) error {
	if c.Message().Private() {
		return c.Send("Регистрация возможна только в групповом чате гильдии.")
	}

	// Проверяем, зарегистрирован ли уже пользователь
	var existingUser models.User
	err := db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", c.Sender().ID).Scan(&existingUser.ID)
	if err == nil {
		return c.Send("Вы уже зарегистрированы.")
	}

	// Разбираем сообщение для регистрации
	// Формат: /register игровой_ник 123456789 название_гильдии роль
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 5 {
		return c.Send("Неверный формат регистрации. Используйте: /register игровой_ник 123456789 название_гильдии роль")
	}

	gameNick := args[1]
	code := args[2]
	guildName := args[3]
	role := strings.ToLower(args[4])

	// Проверяем роль
	validRoles := map[string]bool{"owner": true, "leader": true, "officer": true, "member": true}
	if !validRoles[role] {
		return c.Send("Неверная роль. Допустимые значения: owner, leader, officer, member")
	}

	// Проверяем код (9 цифр)
	if len(code) != 9 {
		return c.Send("Код должен состоять из 9 цифр.")
	}

	// Сохраняем пользователя в базу данных
	_, err = db.Exec(`
		INSERT INTO users (telegram_id, game_nickname, nine_digit_code, guild_name, guild_role, is_active)
		VALUES ($1, $2, $3, $4, $5, TRUE)
	`, c.Sender().ID, gameNick, code, guildName, role)
	if err != nil {
		return c.Send("Ошибка при регистрации. Попробуйте позже.")
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
