package domain

import "errors"

var (
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrInvalidOperation     = errors.New("invalid operation type")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrInternal             = errors.New("internal server error")
)
