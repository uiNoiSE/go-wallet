package service

import (
	"context"
	"fmt"
	"go-wallet/internal/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context) (uuid.UUID, error)
	GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error)
	GetWalletForUpdate(ctx context.Context, id uuid.UUID) (float64, error)
	UpdateWalletBalance(ctx context.Context, id uuid.UUID, newBalance float64) error
	BeginTx(ctx context.Context) (domain.Tx, error)
	WithTx(domain.Tx) WalletRepository
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{
		repo: repo,
	}
}

func (s *WalletService) ProcessTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error {
	if amount <= 0 {
		return domain.ErrAmountMustBePositive
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("service: не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	txRepo := s.repo.WithTx(tx)

	currentBalance, err := txRepo.GetWalletBalance(ctx, id)
	if err != nil {
		return err
	}

	var newBalance float64
	switch opType {
	case domain.OpDeposit:
		newBalance = currentBalance + amount
	case domain.OpWithdraw:
		if currentBalance < amount {
			return domain.ErrInsufficientFunds
		}
		newBalance = currentBalance - amount
	default:
		return domain.ErrInvalidOperation
	}

	err = txRepo.UpdateWalletBalance(ctx, id, newBalance)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("service: не удалось закоммитить транзакцию: %w", err)
	}

	return nil
}

func (s *WalletService) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	return s.repo.CreateWallet(ctx)
}

func (s *WalletService) GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	return s.repo.GetWalletBalance(ctx, id)
}
