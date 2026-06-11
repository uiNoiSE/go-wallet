package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DBConnStr  string
}

func LoadConfig() *Config {
	paths := []string{
		"config.env",
		"../config.env",
		"../../config.env",
		"../../../config.env",
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// postgres://user:password@host:port/db_name?sslmode=sslmode
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=40",
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
