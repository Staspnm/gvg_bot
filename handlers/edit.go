package handlers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gvg-bot/database"
	"github.com/gvg-bot/models"
	"gopkg.in/telebot.v3"
)

// handleEditUserSelf обрабатывает редактирование своих данных
func handleEditUserSelf(c telebot.Context, db *database.Database) error {
	// Получаем текущего пользователя
	var user models.User
	err := db.QueryRow(`
		SELECT id, game_nickname, nine_digit_code, guild_name, guild_role 
		FROM users WHERE telegram_id = $1
	`, c.Sender().ID).Scan(&user.ID, &user.GameNickname, &user.NineDigitCode, &user.GuildName, &user.GuildRole)

	if err != nil {
		return c.Send("❌ Вы не зарегистрированы.")
	}

	// Формат: /editmyinfo новый_ник новый_код
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 3 {
		return c.Send("ℹ️ Используйте: /editmyinfo новый_ник новый_код\nПример: /editmyinfo NewNick 987654321")
	}

	newNick := args[1]
	newCode := args[2]

	// Валидация данных
	if len(newNick) < 3 || len(newNick) > 20 {
		return c.Send("❌ Никнейм должен быть от 3 до 20 символов.")
	}

	if len(newCode) != 9 {
		return c.Send("❌ Код должен состоять из 9 цифр.")
	}

	// Обновляем данные
	_, err = db.Exec(`
		UPDATE users 
		SET game_nickname = $1, nine_digit_code = $2 
		WHERE id = $3
	`, newNick, newCode, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return c.Send("❌ Этот никнейм или код уже заняты.")
		}
		return c.Send("❌ Ошибка при обновлении данных. Попробуйте позже.")
	}

	return c.Send(fmt.Sprintf("✅ Ваши данные успешно обновлены:\nНик: %s\nКод: %s", newNick, newCode))
}

// handleEditUserByOfficer обрабатывает редактирование данных пользователей офицерами
func handleEditUserByOfficer(c telebot.Context, db *database.Database) error {
	// Проверяем права офицера
	var officer models.User
	err := db.QueryRow(`
		SELECT guild_name, guild_role FROM users WHERE telegram_id = $1
	`, c.Sender().ID).Scan(&officer.GuildName, &officer.GuildRole)

	if err != nil {
		return c.Send("❌ Вы не зарегистрированы.")
	}

	if officer.GuildRole != "officer" && officer.GuildRole != "leader" && officer.GuildRole != "owner" {
		return c.Send("❌ У вас нет прав для редактирования данных пользователей.")
	}

	// Формат: /edituser ник_игрока поле новое_значение
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 4 {
		return c.Send("ℹ️ Используйте: /edituser ник_игрока поле новое_значение\n" +
			"Доступные поля: game_nickname, nine_digit_code, guild_name, guild_role\n" +
			"Примеры:\n" +
			"/edituser OldNick game_nickname NewNick\n" +
			"/edituser OldNick nine_digit_code 987654321\n" +
			"/edituser OldNick guild_role officer")
	}

	targetNick := args[1]
	field := args[2]
	newValue := args[3]

	// Проверяем допустимость поля
	validFields := map[string]bool{
		"game_nickname":   true,
		"nine_digit_code": true,
		"guild_name":      true,
		"guild_role":      true,
	}

	if !validFields[field] {
		return c.Send("❌ Недопустимое поле. Доступные: game_nickname, nine_digit_code, guild_name, guild_role")
	}

	// Дополнительные проверки для разных полей
	switch field {
	case "nine_digit_code":
		if len(newValue) != 9 {
			return c.Send("❌ Код должен состоять из 9 цифр.")
		}
	case "guild_role":
		validRoles := map[string]bool{"owner": true, "leader": true, "officer": true, "member": true}
		if !validRoles[newValue] {
			return c.Send("❌ Недопустимая роль. Допустимые: leader, officer, member")
			//return c.Send("❌ Недопустимая роль. Допустимые: owner, leader, officer, member")
		}
		// Ограничения на назначение ролей
		if newValue == "owner" && officer.GuildRole != "owner" {
			return c.Send("❌ Только владелец бота может назначать владельцев.")
		}
		if newValue == "leader" && officer.GuildRole == "member" {
			return c.Send("❌ У вас нет прав назначать лидеров.")
		}
	case "guild_name":
		if officer.GuildRole != "owner" {
			return c.Send("❌ Только владелец бота может изменять гильдию пользователя.")
		}
	}

	// Проверяем существование целевого пользователя
	var targetUser models.User
	err = db.QueryRow(`
		SELECT id, guild_name FROM users 
		WHERE game_nickname = $1 AND guild_name = $2
	`, targetNick, officer.GuildName).Scan(&targetUser.ID, &targetUser.GuildName)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Send(fmt.Sprintf("❌ Пользователь %s не найден в вашей гильдии.", targetNick))
		}
		return c.Send("❌ Ошибка при поиске пользователя.")
	}

	// Особые случаи редактирования
	if field == "guild_name" && officer.GuildRole == "owner" {
		// Владелец может переводить между гильдиями
		_, err = db.Exec(`
			UPDATE users SET guild_name = $1 WHERE id = $2
		`, newValue, targetUser.ID)
	} else if field == "guild_role" {
		// Проверяем права на изменение роли
		if officer.GuildRole == "officer" && (newValue == "leader" || newValue == "owner") {
			return c.Send("❌ Офицер не может назначать лидеров или владельцев.")
		}
		_, err = db.Exec(`
			UPDATE users SET guild_role = $1 WHERE id = $2
		`, newValue, targetUser.ID)
	} else {
		// Обычное поле
		_, err = db.Exec(fmt.Sprintf(`
			UPDATE users SET '%s' = $1 WHERE id = $2
		`, field), newValue, targetUser.ID)
	}

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return c.Send("❌ Это значение уже занято другим пользователем.")
		}
		return c.Send("❌ Ошибка при обновлении данных.")
	}

	return c.Send(fmt.Sprintf("✅ Данные пользователя %s успешно обновлены:\n%s: %s",
		targetNick, field, newValue))
}

// InitEditHandlers инициализирует обработчики редактирования
func InitEditHandlers(bot *telebot.Bot, db *database.Database) {
	bot.Handle("/editmyinfo", func(c telebot.Context) error {
		return handleEditUserSelf(c, db)
	})

	bot.Handle("/edituser", func(c telebot.Context) error {
		return handleEditUserByOfficer(c, db)
	})
}
