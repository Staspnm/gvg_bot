package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gvg-bot/models"
	"strings"

	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

func handleChangeGuild(c telebot.Context, db *database.Database) error {
	// Получаем текущего пользователя
	var user models.User
	err := db.QueryRow(`
        SELECT id, guild_name, guild_role FROM users WHERE telegram_id = $1
    `, c.Sender().ID).Scan(&user.ID, &user.GuildName, &user.GuildRole)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Send("Вы не зарегистрированы. Используйте /register для регистрации.")
		}
		return c.Send("Ошибка при получении ваших данных.")
	}

	// Разбираем команду: /changeguild НовоеНазваниеГильдии
	args := strings.SplitN(c.Message().Text, " ", 2)
	if len(args) < 2 {
		return c.Send("Используйте: /changeguild НовоеНазваниеГильдии\nПример: /changeguild NewDragons")
	}

	newGuildName := strings.TrimSpace(args[1])
	if len(newGuildName) < 2 || len(newGuildName) > 100 {
		return c.Send("Название гильдии должно быть от 2 до 100 символов.")
	}

	// Для обычных участников проверяем, есть ли гильдия с таким названием
	if user.GuildRole == "member" {
		var exists bool
		err := db.QueryRow(`
            SELECT EXISTS(SELECT 1 FROM users WHERE guild_name = $1 LIMIT 1)
        `, newGuildName).Scan(&exists)

		if err != nil || !exists {
			return c.Send("Гильдия с таким названием не найдена. Обратитесь к офицеру.")
		}
	}

	// Обновляем данные
	_, err = db.Exec(`
        UPDATE users SET guild_name = $1 WHERE id = $2
    `, newGuildName, user.ID)

	if err != nil {
		return c.Send("Ошибка при изменении гильдии. Попробуйте позже.")
	}

	// Если пользователь был лидером/офицером, понижаем до участника
	if user.GuildRole == "leader" || user.GuildRole == "officer" {
		_, err = db.Exec(`
            UPDATE users SET guild_role = 'member' WHERE id = $1
        `, user.ID)
		if err != nil {
			return c.Send("Гильдия изменена, но не удалось изменить вашу роль. Обратитесь к администратору.")
		}
		return c.Send(fmt.Sprintf(
			"Вы перешли в гильдию %s. Ваша роль изменена на 'участник'.\n"+
				"Для получения роли офицера обратитесь к лидеру новой гильдии.",
			newGuildName))
	}

	return c.Send(fmt.Sprintf("Вы успешно перешли в гильдию %s!", newGuildName))
}
