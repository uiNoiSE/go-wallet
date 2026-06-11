package service_test

import (
	"context"
	"go-wallet/internal/config"
	"go-wallet/internal/domain"
	"go-wallet/internal/repository"
	"go-wallet/internal/service"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWalletService_ConcurrentTransactions(t *testing.T) {
	cfg := config.LoadConfig()
	pool, err := repository.NewPostgresPool(cfg.DBConnStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	svc := service.NewWalletService(repo)
	ctx := context.Background()
	testWalletPrey, err := svc.CreateWallet(ctx)
	if err != nil {
		t.Fatalf("Не удалось создать тестовый кошелёк: %v", err)
	}

	const (
		concurrentWorkers = 500
		transactionAmount = 10.0
	)

	var wg sync.WaitGroup
	startTime := time.Now()

	var successOps atomic.Int64
	for range concurrentWorkers {
		wg.Go(func() {
			concurrentCtx, concurrentCtxCancel := context.WithTimeout(ctx, 5*time.Second)
			defer concurrentCtxCancel()

			err := svc.ProcessTransaction(concurrentCtx, testWalletPrey, domain.OpDeposit, transactionAmount)
			if err != nil {
				t.Errorf("ошибка при депозите: %v", err)
			} else {
				successOps.Add(1)
			}
		})
	}

	wg.Wait()

	testDuration := time.Since(startTime)
	finalWalletBalance, err := svc.GetBalance(ctx, testWalletPrey)
	if err != nil {
		t.Fatalf("Не удалось получить финальный баланс кошелька :%v", err)
	}

	expectedFinalBalance := float64(successOps.Load()) * transactionAmount
	if finalWalletBalance != expectedFinalBalance {
		t.Errorf("❌ Баланс НЕ сошёлся! Ожидали  %.2f, получили %.2f", expectedFinalBalance, finalWalletBalance)
	} else {
		t.Logf("✅ Баланс кошелька сошёлся: %.2f", finalWalletBalance)
	}

	rps := float64(successOps.Load()) / testDuration.Seconds()
	t.Logf("📊 === МЕТРИКИ НАГРУЗКИ ===")
	t.Logf("Обработано запросов: %d", concurrentWorkers)
	t.Logf("Время выполнения:    %v", testDuration)
	t.Logf("Сила системы (RPS):  %.2f запр/сек", rps)
}
