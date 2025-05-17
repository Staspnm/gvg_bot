package handlers

import (
	"fmt"
	"strings"
	// "time"

	"github.com/gvg-bot/database"
	"github.com/gvg-bot/models"
	"gopkg.in/telebot.v3"
)

func handleBattleReport(c telebot.Context, db *database.Database) error {
	// Проверяем, зарегистрирован ли пользователь и активен ли он
	var user models.User
	err := db.QueryRow(`
		SELECT id, guild_name, is_active FROM users WHERE telegram_id = $1
	`, c.Sender().ID).Scan(&user.ID, &user.GuildName, &user.IsActive)
	if err != nil {
		return c.Send("Вы не зарегистрированы. Используйте /register для регистрации.")
	}

	if !user.IsActive {
		return c.Send("Ваш аккаунт деактивирован. Обратитесь к лидеру гильдии.")
	}

	// Разбираем сообщение с отчетом о битве
	// Формат: T1 вражеский_отряд наш_отряд 15
	parts := strings.Split(c.Message().Text, " ")
	if len(parts) != 4 {
		return c.Send("Неверный формат сообщения. Пример: T1 вражеский_отряд наш_отряд 15")
	}

	location := strings.ToUpper(parts[0])
	enemySquad := parts[1]
	ownSquad := parts[2]
	flagsCount := parts[3]

	// Проверяем локацию
	validLocations := map[string]bool{
		"T1": true, "T2": true, "T3": true, "T4": true,
		"B1": true, "B2": true, "B3": true, "B4": true,
		"F1": true, "F2": true,
	}
	if !validLocations[location] {
		return c.Send("Неверная локация. Допустимые значения: T1-T4, B1-B4, F1-F2")
	}

	// Проверяем количество флагов
	var flags int
	_, err = fmt.Sscanf(flagsCount, "%d", &flags)
	if err != nil || flags < 1 || flags > 22 {
		return c.Send("Количество флагов должно быть числом от 1 до 22")
	}

	// Сохраняем результат битвы
	_, err = db.Exec(`
		INSERT INTO battle_results (user_id, location, enemy_squad, own_squad, flags_count, guild_name)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, location, enemySquad, ownSquad, flags, user.GuildName)
	if err != nil {
		return c.Send("Ошибка при сохранении результата битвы. Попробуйте позже.")
	}

	return c.Send("Результат битвы успешно сохранен!")
}

func handleBattleResults(c telebot.Context, db *database.Database) error {
	// Проверяем, является ли пользователь офицером или выше
	var user models.User
	err := db.QueryRow(`
		SELECT guild_name, guild_role FROM users WHERE telegram_id = $1
	`, c.Sender().ID).Scan(&user.GuildName, &user.GuildRole)
	if err != nil {
		return c.Send("Вы не зарегистрированы.")
	}

	if user.GuildRole != "officer" && user.GuildRole != "leader" && user.GuildRole != "owner" {
		return c.Send("У вас нет прав для просмотра результатов.")
	}

	// Получаем локацию из команды (например, /T1)
	location := strings.ToUpper(strings.TrimPrefix(c.Message().Text, "/"))

	// Получаем результаты для гильдии пользователя
	rows, err := db.Query(`
		SELECT b.location, b.enemy_squad, b.own_squad, b.flags_count, u.game_nickname
		FROM battle_results b
		JOIN users u ON b.user_id = u.id
		WHERE b.guild_name = $1 AND b.location = $2 AND b.battle_date = CURRENT_DATE
		ORDER BY b.reported_at DESC
	`, user.GuildName, location)
	if err != nil {
		return c.Send("Ошибка при получении результатов.")
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var loc, enemy, own, nickname string
		var flags int
		if err := rows.Scan(&loc, &enemy, &own, &flags, &nickname); err != nil {
			continue
		}
		results = append(results, fmt.Sprintf("%s: %s vs %s - %d флагов (от %s)", loc, enemy, own, flags, nickname))
	}

	if len(results) == 0 {
		return c.Send(fmt.Sprintf("Нет результатов для локации %s сегодня.", location))
	}

	response := fmt.Sprintf("Результаты для %s (%s):\n\n%s", location, user.GuildName, strings.Join(results, "\n"))
	return c.Send(response)
}
