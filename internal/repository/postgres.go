package repository

import (
	"context"
	"fmt"
	"go-wallet/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	db   pgxQuerier
	pool *pgxpool.Pool
}

func NewPostgresPool(connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("ERR: Ошибка при создании пула подключений %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("ERR: Пинг пула не прошёл: %w", err)
	}

	return pool, nil
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db:   pool,
		pool: pool,
	}
}

func (r *PostgresRepository) CreateWallet(ctx context.Context) (uuid.UUID, error) {
	newID := uuid.New()
	query := `INSERT INTO wallet_transactions (wallet_id, operation_type, amount) VALUES ($1, 'CREATE', 0.0)`

	_, err := r.db.Exec(ctx, query, newID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Ошибка регистрации кошелька: %w", err)
	}

	return newID, nil
}

func (r *PostgresRepository) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM wallet_transactions WHERE wallet_id = $1)`
	err := r.db.QueryRow(ctx, checkQuery, id).Scan(&exists)
	if err != nil {
		return 0.0, fmt.Errorf("Ошибка проверки существования кошелька: %w", err)
	}
	if !exists {
		return 0.0, fmt.Errorf("Кошелёк с id %s не найден", id)
	}

	query := `
	SELECT COALESCE(
		SUM(CASE WHEN operation_type = 'DEPOSIT' THEN amount ELSE -amount END),
		0.0
	)
	FROM wallet_transactions
	WHERE wallet_id = $1`

	var balance float64
	err = r.db.QueryRow(ctx, query, id).Scan(&balance)
	if err != nil {
		return 0.0, fmt.Errorf("Ошибка при подсчете баланса кошелька: %w", err)
	}

	return balance, nil
}

func (r *PostgresRepository) SaveTransaction(ctx context.Context, id uuid.UUID, opType domain.OperationType, amount float64) error {
	query := `
	INSERT INTO wallet_transactions (wallet_id, operation_type, amount)
	VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, query, id, opType, amount)
	if err != nil {
		return fmt.Errorf("Ошибка при сохранении транзакции в БД: %w", err)
	}

	return nil
}
