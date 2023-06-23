package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	APIKey      string
	ServerPort  string
	LogLevel    string
	// Other configuration parameters
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		APIKey:      os.Getenv("API_KEY"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		// Initialize other configuration parameters from environment variables
	}
}
