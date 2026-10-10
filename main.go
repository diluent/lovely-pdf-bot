package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"lovely-pdf-bot/internal/bot"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warn: .env: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN env var is required")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Bot @%s started", api.Self.UserName)

	if err := bot.RegisterCommands(api); err != nil {
		log.Printf("warn: set commands: %v", err)
	}

	handler := bot.NewHandler(api)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	for upd := range api.GetUpdatesChan(u) {
		if upd.Message == nil {
			continue
		}
		go handler.Handle(upd.Message)
	}
}
