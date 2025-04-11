package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/pkg/logger"
	"time"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	CreateUser(ctx context.Context, user entity.User) (string, error)
}

type postgresUserRepository struct {
	conn   TxManager
	logger logger.Logger
}

func NewUserRepository(conn TxManager, log logger.Logger) UserRepository {
	return &postgresUserRepository{
		conn:   conn,
		logger: log,
	}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user entity.User) (string, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("CreateUser", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)
	query := `
		INSERT INTO users (email, password, role, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	user.CreatedAt = time.Now()

	var userID string
	err := pool.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role, user.CreatedAt).Scan(&userID)
	if err != nil {
		r.logger.Errorw("creating a user",
			"error", err,
			"email", user.Email,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "failed to create user")
	}
	return userID, nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQueryDuration("FindByEmail", time.Since(start).Seconds())
	}()

	pool := r.conn.GetExecutor(ctx)

	query := `
		SELECT id, email, password, role, created_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	err := pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New(errs.ErrNotFoundCode, "user not found")
		}
		r.logger.Errorw("finding a user by email",
			"error", err,
			"email", email,
		)
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to find user by email")
	}

	return &user, nil
}
