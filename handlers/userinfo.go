package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gvg-bot/models"
	"strings"

	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

func handleUserInfo(c telebot.Context, db *database.Database) error {
	// Проверяем, является ли пользователь владельцем
	var user models.User
	err := db.QueryRow(`
        SELECT guild_role FROM users WHERE telegram_id = $1
    `, c.Sender().ID).Scan(&user.GuildRole)

	if err != nil || user.GuildRole != "owner" {
		return c.Send("❌ Эта команда доступна только владельцу бота.")
	}

	// Формат: /userinfo [ник]
	args := strings.Split(c.Message().Text, " ")
	if len(args) < 2 {
		return showAllUsers(c, db)
	}

	nickname := strings.Join(args[1:], " ")
	return showUserDetails(c, db, nickname)
}

func showAllUsers(c telebot.Context, db *database.Database) error {
	rows, err := db.Query(`
        SELECT game_nickname, guild_name, guild_role 
        FROM users 
        ORDER BY guild_name, guild_role DESC, game_nickname
    `)
	if err != nil {
		return c.Send("❌ Ошибка при получении списка пользователей.")
	}
	defer rows.Close()

	var result strings.Builder
	result.WriteString("📊 <b>Список всех пользователей:</b>\n\n")

	currentGuild := ""
	for rows.Next() {
		var nickname, guild, role string
		if err := rows.Scan(&nickname, &guild, &role); err != nil {
			continue
		}

		if guild != currentGuild {
			currentGuild = guild
			result.WriteString(fmt.Sprintf("\n🏰 <b>Гильдия: %s</b>\n", guild))
		}

		roleEmoji := "👤"
		switch role {
		case "owner":
			roleEmoji = "👑"
		case "leader":
			roleEmoji = "⭐"
		case "officer":
			roleEmoji = "🔹"
		}

		result.WriteString(fmt.Sprintf("%s %s (%s)\n", roleEmoji, nickname, role))
	}

	result.WriteString("\nℹ️ Для просмотра подробностей: /userinfo ник")

	return c.Send(result.String(), telebot.ModeHTML)
}

func showUserDetails(c telebot.Context, db *database.Database, nickname string) error {
	var user models.User
	err := db.QueryRow(`
        SELECT game_nickname, nine_digit_code, guild_name, guild_role, is_active, telegram_id
        FROM users WHERE game_nickname = $1
    `, nickname).Scan(&user.GameNickname, &user.NineDigitCode, &user.GuildName, &user.GuildRole, &user.IsActive, &user.TelegramID)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Send(fmt.Sprintf("❌ Пользователь '%s' не найден.", nickname))
		}
		return c.Send("❌ Ошибка при поиске пользователя.")
	}

	activeStatus := "✅ Активен"
	if !user.IsActive {
		activeStatus = "❌ Неактивен"
	}

	roleTitle := map[string]string{
		"owner":   "Владелец бота 👑",
		"leader":  "Лидер гильдии ⭐",
		"officer": "Офицер 🔹",
		"member":  "Участник 👤",
	}[user.GuildRole]

	msg := fmt.Sprintf(`
🔍 <b>Информация о пользователе:</b>

📛 <b>Никнейм:</b> %s
🔢 <b>Код:</b> %s
🏰 <b>Гильдия:</b> %s
🎖️ <b>Роль:</b> %s
%s
🆔 <b>Telegram ID:</b> %d
`,
		user.GameNickname,
		user.NineDigitCode,
		user.GuildName,
		roleTitle,
		activeStatus,
		user.TelegramID)

	// Создаем меню с действиями
	menu := &telebot.ReplyMarkup{}

	// Кнопки для управления пользователем
	if user.GuildRole != "owner" {
		if user.IsActive {
			btnDeactivate := menu.Text("Деактивировать")
			menu.Reply(menu.Row(btnDeactivate))
			c.Set("deactivate_user", user.GameNickname)
		} else {
			btnActivate := menu.Text("Активировать")
			menu.Reply(menu.Row(btnActivate))
			c.Set("activate_user", user.GameNickname)
		}

		if user.GuildRole != "leader" {
			btnPromote := menu.Text("Повысить")
			menu.Reply(menu.Row(btnPromote))
			c.Set("promote_user", user.GameNickname)
		}
	}

	return c.Send(msg, menu, telebot.ModeHTML)
}

// Обработчики кнопок
func InitUserInfoHandlers(bot *telebot.Bot, db *database.Database) {
	bot.Handle(&telebot.Btn{Text: "Деактивировать"}, func(c telebot.Context) error {
		nickname, ok := c.Get("deactivate_user").(string)
		if !ok {
			return c.Send("❌ Не удалось определить пользователя.")
		}

		_, err := db.Exec("UPDATE users SET is_active = FALSE WHERE game_nickname = $1", nickname)
		if err != nil {
			return c.Send("❌ Ошибка при деактивации пользователя.")
		}

		return c.Send(fmt.Sprintf("✅ Пользователь %s деактивирован.", nickname))
	})

	bot.Handle(&telebot.Btn{Text: "Активировать"}, func(c telebot.Context) error {
		nickname, ok := c.Get("activate_user").(string)
		if !ok {
			return c.Send("❌ Не удалось определить пользователя.")
		}

		_, err := db.Exec("UPDATE users SET is_active = TRUE WHERE game_nickname = $1", nickname)
		if err != nil {
			return c.Send("❌ Ошибка при активации пользователя.")
		}

		return c.Send(fmt.Sprintf("✅ Пользователь %s активирован.", nickname))
	})

	bot.Handle(&telebot.Btn{Text: "Повысить"}, func(c telebot.Context) error {
		nickname, ok := c.Get("promote_user").(string)
		if !ok {
			return c.Send("❌ Не удалось определить пользователя.")
		}

		// Определяем текущую роль пользователя
		var currentRole string
		err := db.QueryRow("SELECT guild_role FROM users WHERE game_nickname = $1", nickname).Scan(&currentRole)
		if err != nil {
			return c.Send("❌ Ошибка при проверке роли пользователя.")
		}

		newRole := "officer"
		if currentRole == "member" {
			newRole = "officer"
		} else if currentRole == "officer" {
			newRole = "leader"
		}

		_, err = db.Exec("UPDATE users SET guild_role = $1 WHERE game_nickname = $2", newRole, nickname)
		if err != nil {
			return c.Send("❌ Ошибка при повышении пользователя.")
		}

		return c.Send(fmt.Sprintf("✅ Пользователь %s повышен до %s.", nickname, newRole))
	})
}
