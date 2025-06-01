package registration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gvg-bot/database"

	"github.com/gvg-bot/models"
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

type storage interface {
}

type Usecase struct {
}

func New() *Usecase {
	return &Usecase{}
}

func (u *Usecase) Registration(user User, db *database.Database) error {

	//if c.Message().Private() {
	//	return c.Send("Регистрация возможна только в групповом чате гильдии.")
	//}

	// Проверяем, зарегистрирован ли уже пользователь
	var existingUser models.User
	actualUserID := user.TelegramID
	err := db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", actualUserID).Scan(&existingUser.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("техническая проблема - попробуйте позднее")
		return fmt.Errorf("uscases - registration - db.QueryRow Scan: %w", err)
	}

	if err == nil {
		err = fmt.Errorf("Вы уже зарегистрированы.")
		if err != nil {
			return fmt.Errorf("uscases - registration - db.QueryRow Scan: %w", err)
		}
		return nil
	}

	// Проверяем роль
	validRoles := map[string]bool{"owner": true, "leader": true, "officer": true, "member": true}
	if !validRoles[user.GuildRole] {
		return errors.New("Неверная роль. Допустимые значения: owner, leader, officer, member")
	}

	// Проверяем код (9 цифр)
	if len(user.NineDigitCode) != 9 {
		return errors.New("Код должен состоять из 9 цифр.")
	}

	// Сохраняем пользователя в базу данных
	_, err = db.Exec(`
		INSERT INTO users (telegram_id, game_nickname, nine_digit_code, guild_name, guild_role, is_active)
		VALUES ($1, $2, $3, $4, $5, TRUE)
	`, user.TelegramID, user.GameNickname, user.NineDigitCode, user.GuildName, user.GuildRole)
	if err != nil {
		return fmt.Errorf("Ошибка при регистрации. Попробуйте позже.")
	}
	return nil
}
