package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"queue-bot/utility"
	"strings"
)

type CommandController struct {
	Bot    *tgbotapi.BotAPI
	Update tgbotapi.Update
}

func (c CommandController) ParseCommand() {
	command := c.Update.Message.Command()
	splitString := strings.Split(command, " ")
	args := make([]string, 0, 0)
	if len(splitString) > 1 {
		args = splitString[1:]
	}
	switch command {
	case "show_queues":
		c.ShowQueuesHandler()
		break
	case "create_queue":
		if len(args) == 2 {
			c.CreateQueueHandler(args)
		} else {
			c.Reply(fmt.Sprintf("`/create_queue@%s [name] [lesson]`", c.Bot.Self.UserName))
		}
		break
	}
}

func (c *CommandController) Reply(replyMsg string) {
	msg := tgbotapi.NewMessage(c.Update.Message.Chat.ID, replyMsg)
	msg.ParseMode = "markdown"
	_, err := c.Bot.Send(msg)
	utility.HandleError(err, "Error when replying")
}

func (c *CommandController) ShowQueuesHandler() {

}

func (c *CommandController) CreateQueueHandler(args []string) {
	if len(args) != 2 {
		c.Reply(fmt.Sprintf("`/create_queue@%s [name] [lesson]`", c.Bot.Self.UserName))
	}
}
