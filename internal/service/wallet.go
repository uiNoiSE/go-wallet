package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrInvalidOperation     = errors.New("invalid operation type")
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
		return ErrAmountMustBePositive
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
			return ErrInsufficientFunds
		}
		newBalance = currentBalance - amount
	default:
		return ErrInvalidOperation
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
