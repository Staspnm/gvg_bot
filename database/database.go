package database

import (
	"database/sql"
	"fmt"
	// "log"

	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

func Init(connString string) (*Database, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database - init - db.ping - failed to ping database: %w", err)
	}

	// Применяем миграции
	err = applyMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &Database{db}, nil
}

func applyMigrations(db *sql.DB) error {
	// Здесь можно добавить систему миграций, например, goose или самописную
	// Для простоты будем выполнять SQL-скрипты напрямую

	// Создание таблицы users
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			game_nickname VARCHAR(100) NOT NULL,
			nine_digit_code VARCHAR(9) UNIQUE NOT NULL,
			guild_name VARCHAR(100) NOT NULL,
			guild_role VARCHAR(20) NOT NULL CHECK (guild_role IN ('owner', 'leader', 'officer', 'member')),
			is_active BOOLEAN DEFAULT TRUE,
			registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Создание таблицы battle_results
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS battle_results (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			location VARCHAR(2) NOT NULL CHECK (location IN ('T1', 'T2', 'T3', 'T4', 'B1', 'B2', 'B3', 'B4', 'F1', 'F2')),
			enemy_squad VARCHAR(100) NOT NULL,
			own_squad VARCHAR(100) NOT NULL,
			flags_count INTEGER NOT NULL CHECK (flags_count BETWEEN 1 AND 22),
			reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			guild_name VARCHAR(100) NOT NULL,
			battle_date DATE NOT NULL DEFAULT CURRENT_DATE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create battle_results table: %w", err)
	}

	_, err = db.Exec(`
		ALTER TABLE battle_results
			DROP CONSTRAINT IF EXISTS battle_results_flags_count_check,
			ADD CONSTRAINT battle_results_flags_count_check CHECK (flags_count BETWEEN 0 AND 22)
	`)
	return nil
}
