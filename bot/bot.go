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
