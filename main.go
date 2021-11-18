package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"queue-bot/bot"
	"queue-bot/utility"
)

func init() {
	err := godotenv.Load()
	utility.HandleError(err, "No .env file found")
}

func main() {
	botToken := bot.GetBot()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := botToken.GetUpdatesChan(u)
	utility.HandleError(err, "Errors during updates setup")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		bot.ParseCommand(update)
	}
}
