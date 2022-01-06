package config

import (
	"os"

	"github.com/marcinbor85/pubkey/log"

	"github.com/joho/godotenv"
)

func Get(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.E("cannot find .env file")
	}

	return os.Getenv(key)
}
