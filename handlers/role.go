package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

func handleSetRole(c telebot.Context, db *database.Database) error {
	// Формат: /setrole ник_игрока новая_роль
	args := strings.Split(c.Message().Text, " ")
	if len(args) < 3 {
		return c.Send("ℹ️ Используйте: /setrole ник_игрока роль\nПример: /setrole Player1 leader")
	}

	nickname := args[1]
	newRole := strings.ToLower(args[2])

	// Проверяем валидность роли
	validRoles := map[string]bool{
		"owner":   true,
		"leader":  true,
		"officer": true,
		"member":  true,
	}
	if !validRoles[newRole] {
		return c.Send("❌ Недопустимая роль. Допустимые: owner, leader, officer, member")
	}

	// Находим пользователя
	var userID int
	err := db.QueryRow(`
        SELECT id FROM users WHERE game_nickname = $1
    `, nickname).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Send("❌ Пользователь не найден")
		}
		return c.Send("❌ Ошибка поиска пользователя")
	}

	// Меняем роль
	_, err = db.Exec(`
        UPDATE users SET guild_role = $1 WHERE id = $2
    `, newRole, userID)

	if err != nil {
		log.Printf("Ошибка смены роли: %v", err)
		return c.Send("❌ Ошибка обновления роли")
	}

	return c.Send(fmt.Sprintf(
		"✅ Роль пользователя %s изменена на %s",
		nickname, newRole))
}
