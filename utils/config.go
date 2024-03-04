package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Discord_token string
	DB_URI string
}


func InitConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Fatal("No token found in .env file")
	}

	URI := os.Getenv("DB_URI")
	if URI == "" {
		log.Fatal("No database URI found in .env file")
	}

	return &Config{Discord_token: Token, DB_URI: URI}
}
