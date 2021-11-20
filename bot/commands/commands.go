package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"queue-bot/bot/commands/db"
	"queue-bot/bot/commands/db/models/list"
	"queue-bot/bot/commands/db/models/queue"
	"queue-bot/utility"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxQueues = 5

type CommandController struct {
	Bot    *tgbotapi.BotAPI
	Update tgbotapi.Update
}

func (c *CommandController) BotCanPin() bool {
	bot, err := c.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: c.Update.Message.Chat.ID,
		UserID: c.Bot.Self.ID,
	})
	utility.HandleError(err, "Cannot find bot in chat")
	return bot.CanPinMessages
}

func (c CommandController) ParseCommand() {
	command := c.Update.Message.Command()
	splitString := strings.Split(c.Update.Message.Text, " ")
	args := make([]string, 0, 0)
	if len(splitString) > 1 {
		args = splitString[1:]
	}
	switch command {
	case "show_queues":
		c.ShowQueuesHandler()
		break
	case "create_queue":
		c.CreateQueueHandler(args)
		break
	case "chat_id_dev":
		c.Reply(fmt.Sprintf("%d", c.Update.Message.Chat.ID), "markdown")
		break
	}
}

func (c *CommandController) Reply(replyMsg string, parseMode string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(c.Update.Message.Chat.ID, replyMsg)
	msg.ParseMode = parseMode
	message, err := c.Bot.Send(msg)
	utility.HandleError(err, "Error when replying")
	return message
}

func (c *CommandController) ReplyHtml(replyMsg string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(c.Update.Message.Chat.ID, replyMsg)
	msg.ParseMode = "html"
	message, err := c.Bot.Send(msg)
	utility.HandleError(err, "Error when replying")
	return message
}

func (c *CommandController) ShowQueuesHandler() {
	database := db.ConnectToDb()
	defer database.FinishConnection()
	var queues []queue.Model
	err := database.Tx.Select(&queues, `SELECT * FROM queues WHERE chat_id = ?`, c.Update.Message.Chat.ID)
	utility.HandleError(err, "Error when selecting queues during /show_queues command")
	if len(queues) == 0 {
		c.Reply("Жодної черги в цьому чаті", "markdown")
		return
	}
	response := "`   ID|   Назва| Предмет|`\n"
	maxQ := maxQueues
	for i, v := range queues {
		if i == maxQ {
			maxQ *= 2
			c.Reply(response, "markdown")
			response = "`   ID|   Назва| Предмет|`\n"
		}
		response += "`" + fmt.Sprintf("%5s", strconv.Itoa(v.Id)) + "|" + fmt.Sprintf("%8s", v.Name) + "|" + fmt.Sprintf("%8s", v.Lesson) + "|`\n"
	}
	c.Reply(response, "markdown")
}

func (c *CommandController) CreateQueueHandler(args []string) {
	if len(args) != 2 {
		c.Reply(fmt.Sprintf("/create_queue@%s <i>[назва] [предмет]</i>", c.Bot.Self.UserName), "html")
		return
	}
	if utf8.RuneCountInString(args[0]) > 8 || utf8.RuneCountInString(args[1]) > 8 {
		c.Reply("Не більше восьми символів для назви предмета і черги", "markdown")
		return
	}
	message := c.Reply("Створюємо чергу...", "markdown")
	database := db.ConnectToDb()
	defer database.FinishConnection()
	qModel := queue.Model{
		Name:   args[0],
		Lesson: args[1],
		ChatId: c.Update.Message.Chat.ID,
		MsgId:  message.MessageID,
	}
	id := database.Insert(qModel)
	lastInsertId, err := id.LastInsertId()
	if err != nil {
		utility.HandleError(err, "Error getting last insert id")
		database.Discard()
	}
	lModel := list.Model{
		QueueId:      int(lastInsertId),
		User:         0,
		NumberInList: 0,
	}
	database.Insert(lModel)
	response := "`>>> Черга " + qModel.Name + " <<<`\n"
	response += "`*тут будуть перші 30 людей в черзі*" + "`\n"
	editMsg := tgbotapi.NewEditMessageText(c.Update.Message.Chat.ID, qModel.MsgId, response)
	editMsg.ParseMode = "markdown"
	_, err = c.Bot.Send(editMsg)
	if err != nil {
		utility.HandleError(err, "Error during sending response (queue create)")
		database.Discard()
	}
	if c.BotCanPin() {
		_, err = c.Bot.PinChatMessage(tgbotapi.PinChatMessageConfig{
			ChatID:              c.Update.Message.Chat.ID,
			MessageID:           qModel.MsgId,
			DisableNotification: false,
		})
		if err != nil {
			utility.HandleError(err, "Error during pinning message")
			database.Discard()
		}
	} else {
		c.Reply("Радимо закріпити повідомлення з чергою (або видати для цього права ботові)",
			"markdown")
	}
}
