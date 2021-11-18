package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"queue-bot/utility"
)

func GetBot() *tgbotapi.BotAPI {
	token, isOk := os.LookupEnv("BOT_TOKEN")
	if !isOk {
		log.Fatal("Token is not set")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	utility.HandleError(err, "Error during bot initialization")
	return bot
}

func ParseCommand(update tgbotapi.Update) {
	command := update.Message.Command()
	switch command {
	case "show_queues":
		break
	case "create_queue":
		break
	}
}
