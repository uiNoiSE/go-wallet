package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBConnStr  string
}

func LoadConfig() *Config {
	if err := godotenv.Load("config.env"); err != nil {
		log.Println("WARN: config.env файл не найден, читаем переменные из окружения")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// postgres://user:password@host:port/db_name?sslmode=sslmode
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("DB_SSLMODE"),
	)

	return &Config{
		ServerPort: port,
		DBConnStr:  dbURL,
	}
}
