package main

import (
	"go-wallet/internal/config"
	"go-wallet/internal/handler"
	"go-wallet/internal/repository"
	"go-wallet/internal/service"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := repository.NewPostgresPool(cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Ошибка базы: %v", err)
	}
	defer pool.Close()

	log.Println("🔄 Проверка и запуск миграций Goose...")
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose: не удалось установить диалект: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("goose: ошибка наката миграций: %v", err)
	}
	log.Println("✅ Все миграции успешно применены!")

	log.Println("Пул готов к работе")

	repo := repository.NewPostgresRepository(pool)
	svc := service.NewWalletService(repo)
	h := handler.NewWalletHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", h.CreateWallet)
	mux.HandleFunc("GET /api/v1/wallet/{id}", h.GetBalance)
	mux.HandleFunc("POST /api/v1/wallet/transaction", h.ProcessTransaction)

	serverAddr := ":" + cfg.ServerPort
	log.Printf("Сервер запускается на порту %s...", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, mux))
}
