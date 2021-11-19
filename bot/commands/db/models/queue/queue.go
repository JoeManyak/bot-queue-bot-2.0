package queue

type Model struct {
	Id     int    `db:"id"`
	Name   string `db:"name"`
	Lesson string `db:"lesson"`
	ChatId int64  `db:"chat_id"`
	MsgId  int    `db:"msg_id"`
}

func (m Model) GetFields() []string {
	return []string{"id", "name", "lesson", "chat_id", "msg_id"}
}

func (m Model) GetTable() string {
	return "queues"
}
