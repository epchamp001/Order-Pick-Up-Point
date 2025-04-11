package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"order-pick-up-point/pkg/logger"
)

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxManager interface {
	GetExecutor(ctx context.Context) Executor
	WithTx(ctx context.Context, isoLevel pgx.TxIsoLevel, accessMode pgx.TxAccessMode, fn func(ctx context.Context) error) error
}

var (
	IsolationLevelSerializable   = pgx.Serializable
	IsolationLevelReadCommitted  = pgx.ReadCommitted
	IsolationLevelRepeatableRead = pgx.RepeatableRead

	AccessModeReadWrite = pgx.ReadWrite
	AccessModeReadOnly  = pgx.ReadOnly
)

type txKey struct{}

type postgresTxManager struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

func NewTxManager(pool *pgxpool.Pool, log logger.Logger) TxManager {
	return &postgresTxManager{pool: pool, logger: log}
}

func (t *postgresTxManager) GetExecutor(ctx context.Context) Executor {
	if tx, ok := ctx.Value(txKey{}).(Executor); ok {
		return tx
	}
	return t.pool
}

func (t *postgresTxManager) WithTx(ctx context.Context, isoLevel pgx.TxIsoLevel, accessMode pgx.TxAccessMode, fn func(ctx context.Context) error) error {
	opts := pgx.TxOptions{
		IsoLevel:   isoLevel,
		AccessMode: accessMode,
	}

	var err error

	tx, err := t.pool.BeginTx(ctx, opts)
	if err != nil {
		t.logger.Errorw("Failed to begin transaction",
			"error", err,
		)
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				t.logger.Errorw("Failed to rollback transaction",
					"error", err,
				)
			}
		}

	}()

	ctx = context.WithValue(ctx, txKey{}, tx)
	if err = fn(ctx); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		t.logger.Errorw("Failed to commit transaction",
			"error", err,
		)
	}
	return err
}
