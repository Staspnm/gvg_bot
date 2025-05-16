package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	DBConnString  string
	OwnerID       string
}

func Load() (*Config, error) {
	if err := godotenv.Load("tg.env"); err != nil {
		log.Println("No .env file found")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}

	dbConn := os.Getenv("DB_CONN_STRING")
	if dbConn == "" {
		log.Fatal("DB_CONN_STRING is not set")
	}

	ownerID := ""
	if id := os.Getenv("OWNER_ID"); id != "" {
		ownerID = id
	}

	return &Config{
		TelegramToken: token,
		DBConnString:  dbConn,
		OwnerID:       ownerID,
	}, nil
}

func mustParseInt64(s string) int64 {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		log.Fatalf("Failed to parse number: %v", err)
	}
	return n
}
