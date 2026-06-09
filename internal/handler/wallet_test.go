package handler

import (
	"bytes"
	"context"
	"go-wallet/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type mockWalletService struct {
	onCreateWallet       func(ctx context.Context) (uuid.UUID, error)
	onGetWalletBalance   func(ctx context.Context, id uuid.UUID) (float64, error)
	onProcessTransaction func(ctx context.Context, id uuid.UUID, opType string, amount float64) error
}

func (m *mockWalletService) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	return m.onCreateWallet(ctx)
}

func (m *mockWalletService) GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	return m.onGetWalletBalance(ctx, id)
}

func (m *mockWalletService) ProcessTransaction(ctx context.Context, id uuid.UUID, opType string, amount float64) error {
	return m.onProcessTransaction(ctx, id, opType, amount)
}

func TestCreateWallet(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(m *mockWalletService)
		expectedStatus int
	}{
		{
			name: "Успешное создание кошелька",
			mockBehavior: func(m *mockWalletService) {
				m.onCreateWallet = func(ctx context.Context) (uuid.UUID, error) {
					return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
				}
			},
			expectedStatus: http.StatusCreated,
		},

		{
			name: "Ошибка внутри сервиса",
			mockBehavior: func(m *mockWalletService) {
				m.onCreateWallet = func(ctx context.Context) (uuid.UUID, error) {
					return uuid.Nil, domain.ErrInternal
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSvc := &mockWalletService{}
			test.mockBehavior(mockSvc)
			h := NewWalletHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", nil)
			rec := httptest.NewRecorder()

			h.CreateWallet(rec, req)

			if rec.Code != test.expectedStatus {
				t.Errorf("Ожидался статус %d, получили %d", test.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetWalletBalance(t *testing.T) {
	targetUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name           string
		walletIDPath   string // url
		mockBehavior   func(m *mockWalletService)
		expectedStatus int
	}{
		{
			name:         "Успешное получение баланса",
			walletIDPath: targetUUID.String(),
			mockBehavior: func(m *mockWalletService) {
				m.onGetWalletBalance = func(ctx context.Context, id uuid.UUID) (float64, error) {
					return 150.45, nil
				}
			},
			expectedStatus: http.StatusOK,
		},

		{
			name:         "Кошелёк не найден в базе",
			walletIDPath: targetUUID.String(),
			mockBehavior: func(m *mockWalletService) {
				m.onGetWalletBalance = func(ctx context.Context, id uuid.UUID) (float64, error) {
					return 0, domain.ErrWalletNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},

		{
			name:         "Невалидный формат UUID в URL",
			walletIDPath: "invalid-uuid-101",
			mockBehavior: func(m *mockWalletService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSvc := &mockWalletService{}
			test.mockBehavior(mockSvc)
			h := NewWalletHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/wallet/"+test.walletIDPath, nil)
			req.SetPathValue("id", test.walletIDPath)

			rec := httptest.NewRecorder()

			h.GetWalletBalance(rec, req)
			if rec.Code != test.expectedStatus {
				t.Errorf("Ожидался статус %d, получили %d", test.expectedStatus, rec.Code)
			}
		})
	}
}

func TestProcessTransaction(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockBehavior   func(m *mockWalletService)
		expectedStatus int
	}{
		{
			name:        "Транзакция прошла успешно",
			requestBody: `{"id": "00000000-0000-0000-0000-000000000001", "opType": "DEPOSIT", "amount": 100.0}`,
			mockBehavior: func(m *mockWalletService) {
				m.onProcessTransaction = func(ctx context.Context, id uuid.UUID, opType string, amount float64) error {
					return nil
				}
			},
			expectedStatus: http.StatusOK,
		},

		{
			name:        "Ошибка бизнес-логики (недостаточно средств)",
			requestBody: `{"id": "00000000-0000-0000-0000-000000000001", "opType": "WITHDRAW", "amount": 999999.0}`,
			mockBehavior: func(m *mockWalletService) {
				m.onProcessTransaction = func(ctx context.Context, id uuid.UUID, opType string, amount float64) error {
					return domain.ErrInsufficientFunds
				}
			},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:           "Невалидное тело запроса",
			requestBody:    `some string instead of json`,
			mockBehavior:   nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSvc := &mockWalletService{}
			if test.mockBehavior != nil {
				test.mockBehavior(mockSvc)
			}

			h := NewWalletHandler(mockSvc)

			bodyBuff := bytes.NewBufferString(test.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bodyBuff)
			rec := httptest.NewRecorder()

			h.ProcessTransaction(rec, req)
			if rec.Code != test.expectedStatus {
				t.Errorf("Ожидался статус %d, получили %d", test.expectedStatus, rec.Code)
			}
		})
	}
}
