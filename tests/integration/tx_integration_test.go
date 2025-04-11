//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/logger"
	"time"
)

func (s *TestSuite) TestTransaction_CommitAndRollback() {
	ctx := context.Background()
	log := logger.NewLogger("dev")

	txManager := db.NewTxManager(s.pool, log)

	// Commit: должен сохранить пользователя
	emailCommit := "commit@example.com"
	password := "hashed_password"
	role := "moderator"
	now := time.Now()

	err := txManager.WithTx(ctx, db.IsolationLevelSerializable, db.AccessModeReadWrite, func(ctx context.Context) error {
		executor := txManager.GetExecutor(ctx)
		query := `
			INSERT INTO users (id, email, password, role, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := executor.Exec(ctx, query, uuid.New(), emailCommit, password, role, now)
		return err
	})
	s.Require().NoError(err)

	// Проверяем, что пользователь появился
	dbCheck, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer dbCheck.Close()

	var count int
	err = dbCheck.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", emailCommit).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(1, count, "commit should persist user")

	// Rollback: пользователь НЕ должен быть сохранен
	emailRollback := "rollback@example.com"

	_ = txManager.WithTx(ctx, db.IsolationLevelSerializable, db.AccessModeReadWrite, func(ctx context.Context) error {
		executor := txManager.GetExecutor(ctx)
		query := `
			INSERT INTO users (id, email, password, role, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := executor.Exec(ctx, query, uuid.New(), emailRollback, password, role, now)
		s.Require().NoError(err)

		// Явно эмулируем ошибку
		return fmt.Errorf("simulated failure, should rollback")
	})

	// Проверяем, что rollback сработал - записи нет
	err = dbCheck.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", emailRollback).Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(0, count, "rollback should remove user")
}
