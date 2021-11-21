package commands

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"queue-bot/bot/commands/db"
	"queue-bot/bot/commands/db/models"
	"queue-bot/bot/commands/db/models/list"
	"queue-bot/bot/commands/db/models/queue"
	"queue-bot/utility"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxQueues = 5
const maxListMembers = 10

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
	case "show_queue":
		c.ShowQueueListHandler(args)
		break
	case "show_queues":
		c.ShowQueuesHandler()
		break
	case "create_queue":
		c.CreateQueueHandler(args)
		break
	case "join_queue":
		c.JoinQueueHandler(args)
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
			maxQ += maxQueues
			c.Reply(response, "markdown")
			response = "`   ID|   Назва| Предмет|`\n"
		}
		response += "`" + fmt.Sprintf("%5s", strconv.Itoa(v.Id)) + "|" +
			fmt.Sprintf("%8s", v.Name) + "|" + fmt.Sprintf("%8s", v.Lesson) + "|`\n"
	}
	c.Reply(response, "markdown")
}

func (c *CommandController) CreateQueueHandler(args []string) {
	if len(args) != 2 {
		c.Reply(fmt.Sprintf("/create_queue@%s <i>[назва] [предмет]</i>", c.Bot.Self.UserName),
			"html")
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
			ChatID:    c.Update.Message.Chat.ID,
			MessageID: qModel.MsgId,
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

func (c *CommandController) JoinQueueHandler(args []string) {
	var qId, number = 0, 0
	isValid := true
	if len(args) > 0 {
		var err error
		qId, err = strconv.Atoi(args[0])
		if err != nil {
			isValid = false
		}
		if len(args) > 1 {
			number, err = strconv.Atoi(args[1])
			if err != nil {
				isValid = false
			}
		}
	}

	if (len(args) > 2 || len(args) == 0) || !isValid {
		c.Reply(fmt.Sprintf("/join_queue@%s <i>[id черги] {номер в черзі}</i>", c.Bot.Self.UserName),
			"html")
		return
	}

	if !c.joinQueue(qId, number) {
		return
	}

	c.Reply(fmt.Sprintf("Вас добавлено до черги %d", qId), "markdown")
	c.renewMessage(qId)
}

func (c *CommandController) joinQueue(qId int, number int) bool {
	database := db.ConnectToDb()
	listMap, _, min, _ := c.getQueueList(qId)
	if min > number {
		number = min
	}
	for i := number; ; i++ {
		if listMap[i] == "" {
			lModel := list.Model{
				QueueId:      qId,
				User:         c.Update.Message.From.ID,
				NumberInList: i,
			}
			_, err := database.Tx.NamedExec(models.InsertRequest(lModel), lModel)
			if err != nil {
				database.Discard()
				c.Reply("Такої черги не існує", "markdown")
				return false
			}
			break
		}
	}
	database.FinishConnection()
	return true
}

func (c *CommandController) ShowQueueListHandler(args []string) {
	var qId int
	var err error
	if len(args) > 0 {
		qId, err = strconv.Atoi(args[0])
	}
	if err != nil || len(args) == 0 {
		c.Reply(fmt.Sprintf("/show_queue@%s <i>[id]</i>", c.Bot.Self.UserName), "html")
		return
	}
	qList, _ := c.formQueueList(qId)
	if len(qList) == 0 {
		c.Reply("Такої черги не існує", "markdown")
		return
	}

	response := ""
	maxL := maxListMembers
	for i, v := range qList {
		if i == maxL {
			maxL += maxListMembers
			c.Reply(response, "markdown")
			response = ""
		}
		response += v
	}
	c.Reply(response, "markdown")
}

func (c *CommandController) renewMessage(qId int) {
	queueList, msgId := c.formQueueList(qId)
	result := ""
	for i := 0; i < maxListMembers && i < len(queueList); i++ {
		result += queueList[i]
	}
	msg := tgbotapi.NewEditMessageText(c.Update.Message.Chat.ID, msgId, result)
	msg.ParseMode = "markdown"
	_, err := c.Bot.Send(msg)
	if err != nil {
		c.Reply("Закріплене повідомлення не знайдено", "markdown")
	}
}

func (c *CommandController) formQueueList(qId int) ([]string, int) {
	mapResult, queueName, min, msgId := c.getQueueList(qId)
	result := []string{"`>>> Черга " + queueName + " <<<`\n"}
	for i := min; len(mapResult) != 0; i++ {
		if mapResult[i] == "" {
			result = append(result, "`№"+fmt.Sprintf("%3s", strconv.Itoa(i))+"| Нік: *вільне місце*`\n")
			continue
		}
		result = append(result,
			"`№"+fmt.Sprintf("%3s", strconv.Itoa(i))+"| Нік: "+mapResult[i]+"`\n")
		delete(mapResult, i)
	}
	return result, msgId
}

func (c *CommandController) getQueueList(qId int) (map[int]string, string, int, int) {
	database := db.ConnectToDb()
	defer database.FinishConnection()

	rows, err := database.Tx.Queryx(`SELECT list.number_in_list, list.user, q.name, q.msg_id FROM list JOIN 
	    queues q ON q.id = list.queue_id WHERE q.chat_id = ? AND q.id = ?`, c.Update.Message.Chat.ID, qId)
	utility.HandleError(err, "Error during selecting (formQueueList)")

	var msgId int
	var queueName string
	var min = 1
	mapResult := make(map[int]string)

	for rows.Next() {
		var tempList list.Model
		err = rows.Scan(&tempList.NumberInList, &tempList.User, &queueName, &msgId)
		utility.HandleError(err, "Error during scanning queryx select")
		if tempList.NumberInList == 0 {
			continue
		}
		member, err := c.Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: c.Update.Message.Chat.ID,
			UserID: tempList.User,
		})
		utility.HandleError(err, "Error during finding user "+strconv.Itoa(tempList.User))
		mapResult[tempList.NumberInList] = member.User.UserName
		if min > tempList.NumberInList {
			min = tempList.NumberInList
		}
	}
	return mapResult, queueName, min, msgId
}
