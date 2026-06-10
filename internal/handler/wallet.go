package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go-wallet/internal/domain"
	"net/http"

	"github.com/google/uuid"
)

type WalletService interface {
	CreateWallet(ctx context.Context) (uuid.UUID, error)
	GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error)
	ProcessTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error
}

type WalletHandler struct {
	service WalletService
}

func NewWalletHandler(service WalletService) *WalletHandler {
	return &WalletHandler{
		service: service,
	}
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := h.service.CreateWallet(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		ID string `json:"id"`
	}{ID: id.String()}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid wallet id", http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetWalletBalance(r.Context(), id)
	if err != nil {
		http.Error(w, domain.ErrWalletNotFound.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		Balance float64 `json:"balance"`
	}{Balance: balance}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *WalletHandler) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WalletId      string  `json:"id"`
		OperationType string  `json:"opType"`
		Amount        float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(req.WalletId)
	if err != nil {
		http.Error(w, "Invalid wallet id format", http.StatusBadRequest)
		return
	}

	op := domain.OperationType(req.OperationType)
	if !op.IsValid() {
		http.Error(w, "Invalid operetion type (valid options is: WITHDRAW and DEPOSIT)", http.StatusBadRequest)
		return
	}

	err = h.service.ProcessTransaction(r.Context(), id, op, req.Amount)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WalletHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	// 4xx
	case
		errors.Is(err, domain.ErrAmountMustBePositive),
		errors.Is(err, domain.ErrInsufficientFunds),
		errors.Is(err, domain.ErrInvalidOperation):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, domain.ErrWalletNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	// 5xx
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func respondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
