package main

import (
	"go-wallet/internal/config"
	"go-wallet/internal/repository"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := repository.NewPostgresPool(cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Ошибка базы: %v", err)
	}
	defer pool.Close()

	log.Println("Пул готов к работе")
}
