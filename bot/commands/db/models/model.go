package models

type Model interface {
	GetFields() []string
	GetTable() string
}

func InsertRequest(model Model) string {
	fields := model.GetFields()
	table := model.GetTable()
	list1 := "("
	list2 := "("
	for i := 0; i < len(fields)-1; i++ {
		list1 += fields[i] + ", "
		list2 += ":" + fields[i] + ", "
	}
	list1 += fields[len(fields)-1] + ")"
	list2 += ":" + fields[len(fields)-1] + ")"
	return "INSERT INTO " + table + list1 + " VALUES " + list2
}
