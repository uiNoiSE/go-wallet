package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context) (uuid.UUID, error)
	GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error)
	UpdateWalletBalance(ctx context.Context, id uuid.UUID, newBalance float64) error
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{
		repo: repo,
	}
}

func (s *WalletService) ProcessTransaction(ctx context.Context, id uuid.UUID, opType string, amount float64) error {
	if amount <= 0 {
		err := fmt.Errorf("Сумма транзакции должна быть больше нуля")
		return err
	}
	return nil
}
