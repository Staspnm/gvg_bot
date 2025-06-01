package handlers

import (
	"fmt"
	"strings"

	"github.com/gvg-bot/models"

	"github.com/gvg-bot/database"
	"gopkg.in/telebot.v3"
)

//func handleMissingReports(c telebot.Context, db *database.Database) error {
//	// Проверяем, является ли пользователь офицером или выше
//	var officer models.User
//	err := db.QueryRow(`
//        SELECT guild_name, guild_role FROM users WHERE telegram_id = $1
//    `, c.Sender().ID).Scan(&officer.GuildName, &officer.GuildRole)
//
//	if err != nil {
//		return c.Send("Вы не зарегистрированы.")
//	}
//
//	if officer.GuildRole != "officer" && officer.GuildRole != "leader" && officer.GuildRole != "owner" {
//		return c.Send("У вас нет прав для просмотра этой информации.")
//	}
//
//	// Все возможные локации
//	locations := []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}
//
//	// Получаем список всех активных пользователей гильдии
//	rows, err := db.Query(`
//        SELECT game_nickname
//        FROM users
//        WHERE guild_name = $1 AND is_active = TRUE
//        ORDER BY game_nickname
//    `, officer.GuildName)
//
//	if err != nil {
//		return c.Send("Ошибка при получении списка игроков.")
//	}
//	defer rows.Close()
//
//	var allPlayers []string
//	for rows.Next() {
//		var nickname string
//		if err := rows.Scan(&nickname); err != nil {
//			continue
//		}
//		allPlayers = append(allPlayers, nickname)
//	}
//
//	if len(allPlayers) == 0 {
//		return c.Send("В вашей гильдии нет активных игроков.")
//	}
//
//	// Собираем информацию по каждой локации
//	var result strings.Builder
//	result.WriteString(fmt.Sprintf("Отсутствующие отчеты для гильдии %s:\n\n", officer.GuildName))
//
//	for _, loc := range locations {
//		// Получаем игроков, которые отправили отчеты по этой локации
//		rows, err := db.Query(`
//            SELECT DISTINCT u.game_nickname
//            FROM battle_results b
//            JOIN users u ON b.user_id = u.id
//            WHERE b.guild_name = $1 AND b.location = $2 AND b.battle_date = CURRENT_DATE
//        `, officer.GuildName, loc)
//
//		if err != nil {
//			continue
//		}
//
//		var reportedPlayers []string
//		for rows.Next() {
//			var nickname string
//			if err := rows.Scan(&nickname); err != nil {
//				continue
//			}
//			reportedPlayers = append(reportedPlayers, nickname)
//		}
//		rows.Close()
//
//		// Находим игроков без отчетов
//		missingPlayers := findMissingPlayers(allPlayers, reportedPlayers)
//
//		if len(missingPlayers) > 0 {
//			result.WriteString(fmt.Sprintf("📍 <b>%s</b> (%d):\n", loc, len(missingPlayers)))
//			result.WriteString(strings.Join(missingPlayers, ", "))
//			result.WriteString("\n\n")
//		}
//	}
//
//	if result.Len() > 0 {
//		return c.Send(result.String(), telebot.ModeHTML)
//	}
//
//	return c.Send("Все игроки отправили отчеты по всем локациям!")
//}
//
//func findMissingPlayers(allPlayers, reportedPlayers []string) []string {
//	reportedMap := make(map[string]bool)
//	for _, p := range reportedPlayers {
//		reportedMap[p] = true
//	}
//
//	var missing []string
//	for _, p := range allPlayers {
//		if !reportedMap[p] {
//			missing = append(missing, p)
//		}
//	}
//	return missing
//}

func handleMissingReports(c telebot.Context, db *database.Database, specificLocation string) error {
	// Проверяем права офицера
	var officer models.User
	err := db.QueryRow(`
        SELECT guild_name, guild_role FROM users WHERE telegram_id = $1
    `, c.Sender().ID).Scan(&officer.GuildName, &officer.GuildRole)

	if err != nil {
		return c.Send("Вы не зарегистрированы.")
	}

	if officer.GuildRole != "officer" && officer.GuildRole != "leader" && officer.GuildRole != "owner" {
		return c.Send("У вас нет прав для просмотра этой информации.")
	}

	// Получаем список всех активных пользователей гильдии
	rows, err := db.Query(`
        SELECT game_nickname 
        FROM users 
        WHERE guild_name = $1 AND is_active = TRUE
        ORDER BY game_nickname
    `, officer.GuildName)

	if err != nil {
		return c.Send("Ошибка при получении списка игроков.")
	}
	defer rows.Close()

	var allPlayers []string
	for rows.Next() {
		var nickname string
		if err := rows.Scan(&nickname); err != nil {
			continue
		}
		allPlayers = append(allPlayers, nickname)
	}

	if len(allPlayers) == 0 {
		return c.Send("В вашей гильдии нет активных игроков.")
	}

	// Если указана конкретная локация
	if specificLocation != "" {
		return showMissingForLocation(c, db, officer.GuildName, specificLocation, allPlayers)
	}

	// Иначе показываем по всем локациям
	return showAllMissingReports(c, db, officer.GuildName, allPlayers)
}

func showMissingForLocation(c telebot.Context, db *database.Database, guildName, location string, allPlayers []string) error {
	// Проверяем валидность локации
	validLocations := map[string]bool{
		"T1": true, "T2": true, "T3": true, "T4": true,
		"B1": true, "B2": true, "B3": true, "B4": true,
		"F1": true, "F2": true,
	}

	if !validLocations[location] {
		return c.Send("Недопустимая локация. Допустимые значения: T1-T4, B1-B4, F1-F2")
	}

	// Получаем игроков, отправивших отчеты по этой локации
	rows, err := db.Query(`
        SELECT DISTINCT u.game_nickname
        FROM battle_results b
        JOIN users u ON b.user_id = u.id
        WHERE b.guild_name = $1 AND b.location = $2 AND b.battle_date = CURRENT_DATE
    `, guildName, location)

	if err != nil {
		return c.Send("Ошибка при получении данных.")
	}
	defer rows.Close()

	var reportedPlayers []string
	for rows.Next() {
		var nickname string
		if err := rows.Scan(&nickname); err != nil {
			continue
		}
		reportedPlayers = append(reportedPlayers, nickname)
	}

	// Находим отсутствующие отчеты
	missingPlayers := findMissingPlayers(allPlayers, reportedPlayers)

	if len(missingPlayers) == 0 {
		return c.Send(fmt.Sprintf("Все игроки отправили отчеты по локации %s!", location))
	}

	// Формируем подробный отчет
	var result strings.Builder
	result.WriteString(fmt.Sprintf("📊 <b>Отсутствующие отчеты для %s</b>\n", location))
	result.WriteString(fmt.Sprintf("Гильдия: %s\n", guildName))
	result.WriteString(fmt.Sprintf("Не отчитались (%d):\n\n", len(missingPlayers)))

	for i, player := range missingPlayers {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, player))
	}

	// Добавляем кнопку "Напомнить"
	menu := &telebot.ReplyMarkup{}
	btn := menu.Text("Напомнить игрокам")
	menu.Reply(menu.Row(btn))

	// Сохраняем список игроков для кнопки "Напомнить"
	c.Set("missing_players", missingPlayers)

	return c.Send(result.String(), menu, telebot.ModeHTML)
}

func showAllMissingReports(c telebot.Context, db *database.Database, guildName string, allPlayers []string) error {
	locations := []string{"T1", "T2", "T3", "T4", "B1", "B2", "B3", "B4", "F1", "F2"}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📋 <b>Отсутствующие отчеты для гильдии %s</b>\n\n", guildName))

	anyMissing := false

	for _, loc := range locations {
		rows, err := db.Query(`
            SELECT DISTINCT u.game_nickname
            FROM battle_results b
            JOIN users u ON b.user_id = u.id
            WHERE b.guild_name = $1 AND b.location = $2 AND b.battle_date = CURRENT_DATE
        `, guildName, loc)

		if err != nil {
			continue
		}

		var reportedPlayers []string
		for rows.Next() {
			var nickname string
			if err := rows.Scan(&nickname); err != nil {
				continue
			}
			reportedPlayers = append(reportedPlayers, nickname)
		}
		rows.Close()

		missingPlayers := findMissingPlayers(allPlayers, reportedPlayers)

		if len(missingPlayers) > 0 {
			anyMissing = true
			result.WriteString(fmt.Sprintf("📍 <b>%s</b> (%d):\n", loc, len(missingPlayers)))
			result.WriteString(strings.Join(missingPlayers, ", "))
			result.WriteString("\n\n")
		}
	}

	if !anyMissing {
		return c.Send("🎉 Все игроки отправили отчеты по всем локациям!")
	}

	// Добавляем быстрые кнопки для каждой локации
	menu := &telebot.ReplyMarkup{}
	var buttons []telebot.Btn
	for _, loc := range locations {
		buttons = append(buttons, menu.Text("/missingreports"+loc))
	}
	menu.Reply(menu.Split(3, buttons)...)

	return c.Send(result.String(), menu, telebot.ModeHTML)
}

func findMissingPlayers(allPlayers, reportedPlayers []string) []string {
	reportedMap := make(map[string]bool)
	for _, p := range reportedPlayers {
		reportedMap[p] = true
	}

	var missing []string
	for _, p := range allPlayers {
		isReported := reportedMap[p]
		if !isReported {
			missing = append(missing, p)
		}
	}
	return missing
}
