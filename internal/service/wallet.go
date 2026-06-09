package service

import (
	"context"
	"go-wallet/internal/domain"

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
		return domain.ErrAmountMustBePositive
	}

	currentBalance, err := s.repo.GetWalletBalance(ctx, id)
	if err != nil {
		return err
	}

	var newBalance float64
	switch opType {
	case "DEPOSIT":
		newBalance = currentBalance + amount
	case "WITHDRAW":
		if currentBalance < amount {
			return domain.ErrInsufficientFunds
		}
		newBalance = currentBalance - amount
	default:
		return domain.ErrInvalidOperation
	}

	err = s.repo.UpdateWalletBalance(ctx, id, newBalance)
	if err != nil {
		return err
	}

	return nil
}

func (s *WalletService) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	return s.repo.CreateWallet(ctx)
}

func (s *WalletService) GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	return s.repo.GetWalletBalance(ctx, id)
}
