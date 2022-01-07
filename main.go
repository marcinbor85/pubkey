package main

import (
	"net/http"
	"strconv"

	"github.com/marcinbor85/pubkey/config"
	"github.com/marcinbor85/pubkey/crypto"
	"github.com/marcinbor85/pubkey/database"
	"github.com/marcinbor85/pubkey/tasks"
	eUser "github.com/marcinbor85/pubkey/endpoints/user"
	"github.com/marcinbor85/pubkey/log"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"
)

func main() {
	config.Init(".env")

	lvl, _ := strconv.Atoi(config.Get("LOG_LEVEL"))
	log.Init(log.LogLevel(lvl))

	log.I("application started")

	crypto.Init()

	database.Init(config.Get("DATABASE_FILENAME"))
	defer database.DB.Close()

	c := cron.New()
	c.AddFunc("0 0 0 * * *", func() { tasks.DeleteExpiredRows(database.DB) })
	c.Start()

	router := mux.NewRouter()

	eUser.Register(router)

	http.Handle("/", router)
	http.ListenAndServe("0.0.0.0:"+config.Get("PORT"), nil)
}
