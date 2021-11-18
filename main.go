package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"queue-bot/bot"
	"queue-bot/bot/commands"
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
	controller := commands.CommandController{
		Bot: botToken,
	}
	for update := range updates {
		if update.Message == nil {
			continue
		}
		controller.Update = update
		if update.Message.Command() != "" {
			controller.Reply(controller.Bot.Self.UserName)
		}
		//bot.ParseCommand(update)
	}
}
