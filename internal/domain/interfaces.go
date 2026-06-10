package domain

import "context"

// Tx описывает контракт абстрактной транзакции БД
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
