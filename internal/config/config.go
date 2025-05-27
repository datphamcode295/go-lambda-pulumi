package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	APIKey      string
}

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		APIKey:      os.Getenv("API_KEY"),
	}
}
