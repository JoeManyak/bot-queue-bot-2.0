package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"queue-bot/bot/commands/db/models"
	"queue-bot/utility"
)

type DatabaseController struct {
	Db *sqlx.DB
	Tx *sqlx.Tx
}

func ConnectToDb() DatabaseController {
	login, password, ip, port, dbName := GetDBConnectionArgs()
	db, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s", login, password, ip, port, dbName))
	utility.HandleError(err, "Error during connection to db")
	tx, err := db.Beginx()
	utility.HandleError(err, "Error during starting transaction")
	return DatabaseController{Db: db, Tx: tx}
}

func (d *DatabaseController) Insert(model models.Model) {
	_, err := d.Tx.NamedExec(models.InsertRequest(model), model)
	if err != nil {
		d.Discard()
		utility.HandleError(err, "Error during insertion")
	}
}

func (d *DatabaseController) FinishConnection() {
	err := d.Tx.Commit()
	utility.HandleError(err, "Error during transaction commit")
	err = d.Db.Close()
	utility.HandleError(err, "Error during closing database")
}

func (d *DatabaseController) Discard() {
	err := d.Tx.Rollback()
	utility.HandleError(err, "Error during rollback")
	err = d.Db.Close()
	utility.HandleError(err, "Error during database closing")
}

func GetDBConnectionArgs() (string, string, string, string, string) {
	login, isOk := os.LookupEnv("MYSQL_USER")
	if !isOk {
		log.Fatal("DB login is not set")
	}
	password, isOk := os.LookupEnv("MYSQL_PASSWORD")
	if !isOk {
		log.Fatal("DB password is not set")
	}
	ip, isOk := os.LookupEnv("MYSQL_IP")
	if !isOk {
		log.Fatal("DB ip is not set")
	}
	port, isOk := os.LookupEnv("MYSQL_PORT")
	if !isOk {
		log.Fatal("DB port is not set")
	}
	database, isOk := os.LookupEnv("MYSQL_DATABASE")
	if !isOk {
		log.Fatal("Database port is not set")
	}
	return login, password, ip, port, database
}
