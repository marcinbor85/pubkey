package config

import (
	"os"

	"crypto/rsa"

	r "github.com/marcinbor85/pubkey/crypto/rsa"

	"github.com/joho/godotenv"
)

type Globals struct {
	PrivateKey *rsa.PrivateKey
}

var G = &Globals{}

func initGlobals(g *Globals) error {
	filename := Get("PRIVATE_KEY_FILENAME")
	privateKey, _, err := r.LoadPrivateKey(filename)
	if err != nil {
		return err
	}

	g.PrivateKey = privateKey
	return nil
}

func Init(fileName string) error {
	godotenv.Load(fileName)

	err := initGlobals(G)
	return err
}

func Get(key string) string {
	return os.Getenv(key)
}
