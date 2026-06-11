package service

import (
	"context"
	"errors"
	"go-wallet/internal/domain"
	"testing"

	"github.com/google/uuid"
)

type mockRepo struct {
	balance float64
	dbErr   error
}

func (m *mockRepo) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockRepo) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	if m.dbErr != nil {
		return 0, m.dbErr
	}
	return m.balance, nil
}

func (m *mockRepo) SaveTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error {
	if m.dbErr != nil {
		return m.dbErr
	}

	if opType == domain.OpDeposit {
		m.balance += amount
	} else {
		m.balance -= amount
	}

	return nil
}

func TestProcessTransaction(t *testing.T) {
	type testCase struct {
		name         string
		initialFunds float64
		opType       domain.OperationType
		amount       float64
		dbErr        error
		wantNewFunds float64
		wantErr      error
	}

	tests := []testCase{
		{
			name:         "Успешный депозит",
			initialFunds: 100.0,
			opType:       domain.OpDeposit,
			amount:       50.0,
			wantNewFunds: 150.0,
			wantErr:      nil,
		},

		{
			name:         "Успешное списание",
			initialFunds: 100.0,
			opType:       domain.OpWithdraw,
			amount:       40.0,
			wantNewFunds: 60.0,
			wantErr:      nil,
		},

		{
			name:         "Ошибка: Сумма транзакции меньше или равна нулю",
			initialFunds: 100.0,
			opType:       domain.OpDeposit,
			amount:       -10.0,
			wantNewFunds: 100.0,
			wantErr:      domain.ErrAmountMustBePositive,
		},

		{
			name:         "Ошибка: Недостаточно средств",
			initialFunds: 50.0,
			opType:       domain.OpWithdraw,
			amount:       100.0,
			wantNewFunds: 50.0,
			wantErr:      domain.ErrInsufficientFunds,
		},

		{
			name:         "Ошибка: Неверный тип операции",
			initialFunds: 100.0,
			opType:       "TRANSFER",
			amount:       20.0,
			wantNewFunds: 100.0,
			wantErr:      domain.ErrInvalidOperation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeDB := &mockRepo{balance: tc.initialFunds, dbErr: tc.dbErr}
			svc := NewWalletService(fakeDB)

			walletID := uuid.New()
			err := svc.ProcessTransaction(context.Background(), walletID, tc.opType, tc.amount)

			if tc.wantErr != nil {

				if err == nil {
					t.Fatalf("Ожидали ошибку '%v', но метод выполнился успешно", tc.wantErr)
				}

				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Ожидали ошибку '%v', но получили '%v'", tc.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Не ожидали ошибку, но получили: %v", err)
				}
			}

			if fakeDB.balance != tc.wantNewFunds {
				t.Errorf("Ошибка баланса! Ожидали финальный баланс %.2f, но получили %.2f", tc.wantNewFunds, fakeDB.balance)
			}
		})
	}
}
