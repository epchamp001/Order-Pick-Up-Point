package http

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/models/entity"
	"order-pick-up-point/internal/storage/db"
	"order-pick-up-point/pkg/jwt"
	"order-pick-up-point/pkg/logger"
	"order-pick-up-point/pkg/password"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, email string, password string, role string) (string, error)
	Login(ctx context.Context, email string, password string) (string, error)
	DummyLogin(ctx context.Context, role string) (string, error)
}

type authServiceImp struct {
	repo         db.Repository
	txManager    db.TxManager
	tokenSvc     jwt.TokenService
	hasher       password.PasswordHasher
	logger       logger.Logger
	allowedRoles map[string]bool
}

func NewAuthService(
	repo db.Repository,
	txManager db.TxManager,
	tokenSvc jwt.TokenService,
	hasher password.PasswordHasher,
	logger logger.Logger,
	roles map[string]bool,
) AuthService {
	return &authServiceImp{
		repo:         repo,
		txManager:    txManager,
		tokenSvc:     tokenSvc,
		hasher:       hasher,
		logger:       logger,
		allowedRoles: roles,
	}
}

func (s *authServiceImp) Register(ctx context.Context, email string, passwordStr string, role string) (string, error) {
	if err := s.validateRole(role); err != nil {
		s.logger.Errorw("Register",
			"role", role,
			"email", email,
			"error", err,
		)
		return "", err
	}

	var userID string

	err := s.txManager.WithTx(ctx, pgx.ReadCommitted, pgx.ReadWrite, func(txCtx context.Context) error {
		existingUser, err := s.repo.FindByEmail(txCtx, email)
		if existingUser != nil {
			return errs.New(errs.ErrUserAlreadyExists, "user already exists")
		}

		if err != nil && !errs.IsNotFound(err) {
			return err
		}

		hashedPassword, err := s.hasher.Hash(passwordStr)
		if err != nil {
			s.logger.Errorw("hash password",
				"error", err,
				"email", email,
			)

			return errs.Wrap(err, errs.ErrPasswordHashingFailed, "failed to hash password")
		}

		user := entity.User{
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         role,
			CreatedAt:    time.Now(),
		}

		uid, err := s.repo.CreateUser(txCtx, user)
		if err != nil {
			return err
		}
		userID = uid
		return nil
	})
	if err != nil {
		s.logger.Errorw("Register",
			"error", err,
			"email", email,
		)
		return "", err
	}

	return userID, nil
}

func (s *authServiceImp) Login(ctx context.Context, email string, passwordStr string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Errorw("Login",
			"email", email,
			"error", err,
		)
		return "", errs.New(errs.ErrInvalidCredentials, "invalid credentials")
	}

	if !s.hasher.Check(user.PasswordHash, passwordStr) {
		errCheckPass := errs.New(errs.ErrInvalidCredentials, "invalid credentials")
		s.logger.Errorw("Login",
			"email", email,
			"error", errCheckPass,
		)
		return "", errCheckPass
	}

	tokenStr, err := s.tokenSvc.GenerateToken(user.ID, user.Role)
	if err != nil {
		s.logger.Errorw("Login",
			"error", err,
			"userID", user.ID,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "failed to generate token")
	}

	return tokenStr, nil
}

func (s *authServiceImp) DummyLogin(ctx context.Context, role string) (string, error) {
	if err := s.validateRole(role); err != nil {
		s.logger.Errorw("DummyLogin",
			"role", role,
			"error", err,
		)
		return "", err
	}

	tokenStr, err := s.tokenSvc.GenerateToken("dummyID", role)
	if err != nil {
		s.logger.Errorw("DummyLogin",
			"role", role,
			"error", err,
		)
		return "", errs.Wrap(err, errs.ErrInternalCode, "dummy login failed")
	}
	return tokenStr, nil
}

func (s *authServiceImp) validateRole(role string) error {
	if !s.allowedRoles[role] {
		return errs.New(errs.ErrInvalidRoleCode, fmt.Sprintf("role '%s' is not allowed", role))
	}
	return nil
}
