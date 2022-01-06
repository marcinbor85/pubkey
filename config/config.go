package config

import (
	"os"

	"github.com/joho/godotenv"
)

func Init(fileName string) {
	godotenv.Load(fileName)
}

func Get(key string) string {
	return os.Getenv(key)
}
