package repository

import (
	"context"
	"fmt"

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

	query := `INSERT INTO wallets (id, balance) VALUES ($1, $2)`

	_, err := r.db.Exec(ctx, query, newID, 0.00)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Ошибка при вставке кошелька в БД: %w", err)
	}

	return newID, nil
}

func (r *PostgresRepository) GetWalletBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	var balance float64

	query := `SELECT balance FROM wallets WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&balance)
	if err != nil {
		return 0.0, fmt.Errorf("Ошибка при получении баланса кошелька: %w", err)
	}

	return balance, nil
}

func (r *PostgresRepository) UpdateWalletBalance(ctx context.Context, id uuid.UUID, newBalance float64) error {
	query := `UPDATE wallets SET balance = $1 WHERE id = $2`

	result, err := r.db.Exec(ctx, query, newBalance, id)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении баланса в БД: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Кошелёк с id: %s не найден", id)
	}

	return nil
}

func (r *PostgresRepository) WithTx(tx pgxQuerier) *PostgresRepository {
	return &PostgresRepository{
		db: tx,
	}
}

func (r *PostgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("Не удалось открыть транзакцию: %w", err)
	}
	return tx, nil
}
