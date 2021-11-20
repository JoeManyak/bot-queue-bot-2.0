package list

type Model struct {
	Id           int `db:"id"`
	QueueId      int `db:"queue_id"`
	User         int `db:"user"`
	NumberInList int `db:"number_in_list"`
}

func (m Model) GetFields() []string {
	return []string{"id", "queue_id", "user", "number_in_list"}
}

func (m Model) GetTable() string {
	return "list"
}
