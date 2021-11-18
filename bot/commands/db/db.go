package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"queue-bot/utility"
)

type DatabaseController struct {
	Db *sqlx.DB
}

func ConnectToDb() DatabaseController {
	login, password, ip, port, dbName := GetDBConnectionArgs()
	db, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@(%s:%s)/%s", login, password, ip, port, dbName))
	utility.HandleError(err, "Error during connection to db")
	return DatabaseController{Db: db}
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
