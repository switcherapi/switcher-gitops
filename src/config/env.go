package config

import (
	"os"

	"github.com/joho/godotenv"
)

func InitEnv() {
	goEnv := os.Getenv("GO_ENV")
	godotenv.Load("../../.env." + goEnv)
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
