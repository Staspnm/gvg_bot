package handlers

import (
	"fmt"
	"github.com/gvg-bot/models"
	"strings"

	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

func handleEditUserSelf(c telebot.Context, db *database.Database) error {
	// Получаем текущего пользователя
	var user models.User
	err := db.QueryRow(`
        SELECT id, game_nickname, nine_digit_code, guild_name, guild_role 
        FROM users WHERE telegram_id = $1
    `, c.Sender().ID).Scan(&user.ID, &user.GameNickname, &user.NineDigitCode, &user.GuildName, &user.GuildRole)

	if err != nil {
		return c.Send("Вы не зарегистрированы.")
	}

	// Формат: /editmyinfo новый_ник новый_код
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 3 {
		return c.Send("Используйте: /editmyinfo новый_ник новый_код\nПример: /editmyinfo NewNick 987654321")
	}

	newNick := args[1]
	newCode := args[2]

	// Проверяем код (9 цифр)
	if len(newCode) != 9 {
		return c.Send("Код должен состоять из 9 цифр.")
	}

	// Обновляем данные
	_, err = db.Exec(`
        UPDATE users 
        SET game_nickname = $1, nine_digit_code = $2 
        WHERE id = $3
    `, newNick, newCode, user.ID)

	if err != nil {
		return c.Send("Ошибка при обновлении данных. Попробуйте позже.")
	}

	return c.Send("Ваши данные успешно обновлены!")
}

func handleEditUserByOfficer(c telebot.Context, db *database.Database) error {
	// Проверяем, является ли пользователь офицером или выше
	var officer models.User
	err := db.QueryRow(`
        SELECT guild_name, guild_role FROM users WHERE telegram_id = $1
    `, c.Sender().ID).Scan(&officer.GuildName, &officer.GuildRole)

	if err != nil {
		return c.Send("Вы не зарегистрированы.")
	}

	if officer.GuildRole != "officer" && officer.GuildRole != "leader" && officer.GuildRole != "owner" {
		return c.Send("У вас нет прав для редактирования данных пользователей.")
	}

	// Формат: /edituser ник_игрока поле новое_значение
	// Пример: /edituser OldNick game_nickname NewNick
	// Или: /edituser OldNick nine_digit_code 123456789
	args := strings.Split(c.Message().Text, " ")
	if len(args) != 4 {
		return c.Send("Используйте: /edituser ник_игрока поле новое_значение\n" +
			"Доступные поля: game_nickname, nine_digit_code, guild_name, guild_role\n" +
			"Примеры:\n" +
			"/edituser OldNick game_nickname NewNick\n" +
			"/edituser OldNick nine_digit_code 987654321\n" +
			"/edituser OldNick guild_role officer")
	}

	targetNick := args[1]
	field := args[2]
	newValue := args[3]

	// Проверяем допустимость поля для редактирования
	validFields := map[string]bool{
		"game_nickname":   true,
		"nine_digit_code": true,
		"guild_name":      true,
		"guild_role":      true,
	}

	if !validFields[field] {
		return c.Send("Недопустимое поле для редактирования. Доступные поля: game_nickname, nine_digit_code, guild_name, guild_role")
	}

	// Для кода проверяем длину
	if field == "nine_digit_code" && len(newValue) != 9 {
		return c.Send("Код должен состоять из 9 цифр.")
	}

	// Для роли проверяем допустимые значения
	if field == "guild_role" {
		validRoles := map[string]bool{"owner": true, "leader": true, "officer": true, "member": true}
		if !validRoles[newValue] {
			return c.Send("Недопустимая роль. Допустимые значения: owner, leader, officer, member")
		}

		// Офицер не может назначать роль выше своей
		if (newValue == "owner" || newValue == "leader") && officer.GuildRole != "owner" {
			return c.Send("Вы не можете назначать эту роль.")
		}
	}

	// Обновляем данные
	_, err = db.Exec(fmt.Sprintf(`
        UPDATE users 
        SET %s = $1 
        WHERE game_nickname = $2 AND guild_name = $3
    `, field), newValue, targetNick, officer.GuildName)

	if err != nil {
		return c.Send("Ошибка при обновлении данных. Возможно, пользователь не найден.")
	}

	return c.Send(fmt.Sprintf("Данные пользователя %s успешно обновлены!", targetNick))
}
