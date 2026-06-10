package main

import (
	"go-wallet/internal/config"
	"go-wallet/internal/handler"
	"go-wallet/internal/repository"
	"go-wallet/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := repository.NewPostgresPool(cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Ошибка базы: %v", err)
	}
	defer pool.Close()

	log.Println("Пул готов к работе")

	repo := repository.NewPostgresRepository(pool)
	svc := service.NewWalletService(repo)
	h := handler.NewWalletHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", h.CreateWallet)
	mux.HandleFunc("GET /api/v1/wallet/{id}", h.GetWalletBalance)
	mux.HandleFunc("POST /api/v1/wallet/transaction", h.ProcessTransaction)

	serverAddr := ":" + cfg.ServerPort
	log.Printf("Сервер запускается на порту %s...", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, mux))
}
