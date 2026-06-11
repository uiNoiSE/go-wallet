package service

import (
	"context"
	"fmt"
	"go-wallet/internal/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	CreateWallet(ctx context.Context) (uuid.UUID, error)
	GetBalance(ctx context.Context, id uuid.UUID) (float64, error)
	SaveTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	return s.repo.CreateWallet(ctx)
}

func (s *WalletService) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	return s.repo.GetBalance(ctx, id)
}

func (s *WalletService) ProcessTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error {
	if opType != domain.OpDeposit && opType != domain.OpWithdraw {
		return domain.ErrInvalidOperation
	}

	if amount <= 0 {
		return domain.ErrAmountMustBePositive
	}

	if opType == domain.OpWithdraw {
		currentBalance, err := s.repo.GetBalance(ctx, id)
		if err != nil {
			return fmt.Errorf("ошибка при проверке баланса: %w", err)
		}
		if currentBalance < amount {
			return domain.ErrInsufficientFunds
		}
	}

	err := s.repo.SaveTransaction(ctx, id, opType, amount)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении операции: %w", err)
	}

	return nil
}
