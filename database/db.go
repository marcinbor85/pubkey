package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"

	"github.com/marcinbor85/pubkey/log"
	mUser "github.com/marcinbor85/pubkey/models/user"
)

var DB *sql.DB

func Init(filename string) {
	log.D(`database filename "%s"`, filename)

	new_database := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.W("no database")
		file, err := os.Create(filename)
		if err != nil {
			panic(err.Error())
		}
		file.Close()
		log.I("database created")
		new_database = true
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		panic(err.Error())
	}

	DB = db
	log.I("database opened")

	if new_database != false {
		mUser.CreateTable(DB)

		log.I("models created")
	}
}
