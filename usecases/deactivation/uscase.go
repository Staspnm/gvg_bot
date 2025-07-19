package deactivation

import (
	"fmt"

	"github.com/gvg-bot/database"
)

type User struct {
	TelegramID    int64
	DiscordID     string
	GameNickname  string
	NineDigitCode string
	GuildName     string
	GuildRole     string // owner, leader, officer, member
	IsActive      bool
}

type Usecase struct {
}

func New() *Usecase {
	return &Usecase{}
}

func (u *Usecase) Deactivation(user User, db *database.Database) error {

	// Проверяем права (только owner/leader могут деактивировать)
	var requesterRole string
	err := db.QueryRow("SELECT guild_role FROM users WHERE telegram_id = $1", user.TelegramID).Scan(&requesterRole)
	if err != nil {
		return fmt.Errorf("Вы не зарегистрированы.")
	}

	if requesterRole != "owner" && requesterRole != "leader" {

		return fmt.Errorf("У вас нет прав для этой команды.")

	}

	// Получаем ID пользователя для деактивации (из reply или упоминания)
	targetUser := c.Message().ReplyTo.Sender
	if targetUser == nil {
		return c.Send("Ответьте на сообщение пользователя или упомяните его.")
	}

	// Деактивируем пользователя
	_, err = db.Exec("UPDATE users SET is_active = FALSE WHERE telegram_id = $1", user.TelegramID)
	if err != nil {
		return fmt.Errorf("Ошибка при деактивации пользователя.")
	}

	return c.Send(fmt.Sprintf("Пользователь @%s деактивирован.", targetUser.Username))
}
