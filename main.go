import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

package main

import (
"log"
"os"
"os/signal"
"syscall"

"github.com/gvg-bot/config"
"github.com/gvg-bot/database"
"github.com/gvg-bot/handlers"
"gopkg.in/telebot.v3"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация базы данных
	db, err := database.Init(cfg.DBConnString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Настройка бота
	botSettings := telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(botSettings)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Инициализация обработчиков
	handlers.InitHandlers(bot, db)

	// Запуск бота
	go bot.Start()

	// Ожидание сигнала для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down bot...")
}