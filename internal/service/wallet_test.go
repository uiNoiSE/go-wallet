package service

import (
	"context"
	"errors"
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

func (m *mockRepo) GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	if m.dbErr != nil {
		return 0, m.dbErr
	}
	return m.balance, nil
}

func (m *mockRepo) UpdateWalletBalance(ctx context.Context, id uuid.UUID, newBalance float64) error {
	if m.dbErr != nil {
		return m.dbErr
	}

	m.balance = newBalance
	return nil
}

func TestProcessTransaction(t *testing.T) {
	type testCase struct {
		name         string
		initialFunds float64
		opType       string
		amount       float64
		dbErr        error
		wantNewFunds float64
		wantErr      error
	}

	tests := []testCase{
		{
			name:         "Успешный депозит",
			initialFunds: 100.0,
			opType:       "DEPOSIT",
			amount:       50.0,
			wantNewFunds: 150.0,
			wantErr:      nil,
		},

		{
			name:         "Успешное списание",
			initialFunds: 100.0,
			opType:       "WITHDRAW",
			amount:       40.0,
			wantNewFunds: 60.0,
			wantErr:      nil,
		},

		{
			name:         "Ошибка: Сумма транзакции меньше или равна нулю",
			initialFunds: 100.0,
			opType:       "DEPOSIT",
			amount:       -10.0,
			wantNewFunds: 100.0,
			wantErr:      ErrAmountMustBePositive,
		},

		{
			name:         "Ошибка: Недостаточно средств",
			initialFunds: 50.0,
			opType:       "WITHDRAW",
			amount:       100.0,
			wantNewFunds: 50.0,
			wantErr:      ErrInsufficientFunds,
		},

		{
			name:         "Ошибка: Неверный тип операции",
			initialFunds: 100.0,
			opType:       "TRANSFER",
			amount:       20.0,
			wantNewFunds: 100.0,
			wantErr:      ErrInvalidOperation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeDB := &mockRepo{balance: tc.initialFunds, dbErr: tc.dbErr}
			svc := NewWalletService(fakeDB)

			err := svc.ProcessTransaction(context.Background(), uuid.New(), tc.opType, tc.amount)

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
		})
	}
}
