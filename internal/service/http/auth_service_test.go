package http

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/models/entity"
	mockRepo "order-pick-up-point/internal/storage/db/mock"
	mockToken "order-pick-up-point/pkg/jwt/mock"
	mockLog "order-pick-up-point/pkg/logger/mock"
	mockPass "order-pick-up-point/pkg/password/mock"
	"strings"
	"testing"
)

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"
	passwordStr := "secret"
	validRole := "client"
	invalidRole := "unknown"
	fakeUserID := "user123"

	allowedRoles := map[string]bool{
		"client":    true,
		"moderator": true,
	}

	tests := []struct {
		name           string
		email          string
		password       string
		role           string
		allowedRoles   map[string]bool
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "invalid role",
			email:          email,
			password:       passwordStr,
			role:           invalidRole,
			allowedRoles:   allowedRoles,
			expectedErrMsg: "role 'unknown' is not allowed",
		},
		{
			name:           "tx error",
			email:          email,
			password:       passwordStr,
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "user already exists",
			email:          email,
			password:       passwordStr,
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "existing",
			expectedErrMsg: "user already exists",
		},
		{
			name:           "error in find user",
			email:          email,
			password:       passwordStr,
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "find",
			expectedErrMsg: "find error",
		},
		{
			name:           "hashing error",
			email:          email,
			password:       passwordStr,
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "hash",
			expectedErrMsg: "failed to hash password",
		},
		{
			name:           "error in CreateUser",
			email:          email,
			password:       passwordStr,
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "create",
			expectedErrMsg: "create error",
		},
		{
			name:         "success",
			email:        email,
			password:     passwordStr,
			role:         validRole,
			allowedRoles: allowedRoles,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)
			hasherMock := mockPass.NewPasswordHasher(t)

			svc := &authServiceImp{
				repo:         repoMock,
				txManager:    txManager,
				hasher:       hasherMock,
				logger:       loggerMock,
				allowedRoles: tc.allowedRoles,
			}

			// Если роль недопустима, validateRole вернет ошибку до вызова транзакции
			if tc.expectedErrMsg != "" && strings.Contains(tc.expectedErrMsg, "not allowed") {
				loggerMock.
					On("Errorw", "Register", "role", tc.role, "email", tc.email, "error", mock.Anything).
					Return().
					Once()
				uid, err := svc.Register(ctx, tc.email, tc.password, tc.role)
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v (uid: %q)", tc.expectedErrMsg, err, uid)
				}
				return
			}

			var callbackErr error

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Return(errors.New("tx error")).
					Once()
				loggerMock.
					On("Errorw", "Register", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					}), "email", tc.email).
					Return().
					Once()
			} else {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Run(func(args mock.Arguments) {
						f := args.Get(3).(func(context.Context) error)
						callbackErr = f(ctx)
					}).
					Return(func(_ context.Context, _ pgx.TxIsoLevel, _ pgx.TxAccessMode, _ func(context.Context) error) error {
						return callbackErr
					}).
					Once()

				switch tc.simulateError {
				case "existing":
					existingUser := &entity.User{Email: tc.email}
					repoMock.
						On("FindByEmail", mock.Anything, tc.email).
						Return(existingUser, nil).
						Once()
				case "find":
					repoMock.
						On("FindByEmail", mock.Anything, tc.email).
						Return(nil, errors.New("find error")).
						Once()
				default:
					// Ожидаем, что пользователь не найден
					repoMock.
						On("FindByEmail", mock.Anything, tc.email).
						Return(nil, errs.New(errs.ErrNotFoundCode, "not found")).
						Once()
				}

				// Если пользователь не найден – вызывается hasher.Hash
				if tc.simulateError == "hash" {
					hasherMock.
						On("Hash", tc.password).
						Return("", errors.New("hash error")).
						Once()
					loggerMock.
						On("Errorw", "hash password", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), "hash error")
						}), "email", tc.email).
						Return().
						Once()
				} else if tc.simulateError == "create" {
					// В сценарии "create" хэширование проходит успешно
					hasherMock.
						On("Hash", tc.password).
						Return("hashed_secret", nil).
						Once()
				} else if tc.simulateError == "" || tc.simulateError == "existing" || tc.simulateError == "find" {
					hasherMock.
						On("Hash", tc.password).
						Return("hashed_secret", nil).
						Maybe()
				}

				// Если хэширование прошло успешно, вызывается CreateUser
				if tc.simulateError == "create" {
					repoMock.
						On("CreateUser", mock.Anything, mock.MatchedBy(func(u entity.User) bool {
							return u.Email == tc.email && u.PasswordHash == "hashed_secret" && u.Role == tc.role
						})).
						Return("", errors.New("create error")).
						Once()
				} else if tc.simulateError == "" {
					repoMock.
						On("CreateUser", mock.Anything, mock.MatchedBy(func(u entity.User) bool {
							return u.Email == tc.email && u.PasswordHash == "hashed_secret" && u.Role == tc.role
						})).
						Return(fakeUserID, nil).
						Once()
				}

				if tc.expectedErrMsg != "" && tc.simulateError != "tx" {
					loggerMock.
						On("Errorw", "Register", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						}), "email", tc.email).
						Return().
						Once()
				}
			}

			userID, err := svc.Register(ctx, tc.email, tc.password, tc.role)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && userID != fakeUserID {
				t.Errorf("expected userID %q, got %q", fakeUserID, userID)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"
	passwordStr := "secret"

	user := &entity.User{
		ID:           "user123",
		Email:        email,
		PasswordHash: "hashed_secret",
		Role:         "client",
	}

	fakeToken := "jwt_token_abc"

	tests := []struct {
		name           string
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "error in find user",
			simulateError:  "find",
			expectedErrMsg: "invalid credentials",
		},
		{
			name:           "wrong password",
			simulateError:  "wrong_pass",
			expectedErrMsg: "invalid credentials",
		},
		{
			name:           "error in token generate",
			simulateError:  "token",
			expectedErrMsg: "failed to generate token",
		},
		{
			name:           "success",
			simulateError:  "",
			expectedErrMsg: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			tokenSvcMock := mockToken.NewTokenService(t)
			hasherMock := mockPass.NewPasswordHasher(t)

			svc := &authServiceImp{
				repo: repoMock,
				// txManager не нужен для Login
				tokenSvc:     tokenSvcMock,
				hasher:       hasherMock,
				logger:       loggerMock,
				allowedRoles: map[string]bool{"client": true, "moderator": true},
			}

			switch tc.simulateError {
			case "find":
				// repo.FindByEmail возвращает ошибку
				repoMock.
					On("FindByEmail", mock.Anything, email).
					Return(nil, errors.New("db error")).
					Once()
				loggerMock.
					On("Errorw", "Login", "email", email, "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "db error")
					})).
					Return().
					Once()
			case "wrong_pass":
				// Пользователь найден, но hasher.Check возвращает false
				repoMock.
					On("FindByEmail", mock.Anything, email).
					Return(user, nil).
					Once()
				hasherMock.
					On("Check", user.PasswordHash, passwordStr).
					Return(false).
					Once()
				loggerMock.
					On("Errorw", "Login", "email", email, "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "invalid credentials")
					})).
					Return().
					Once()
			case "token":
				// Пользователь найден, пароль верный
				repoMock.
					On("FindByEmail", mock.Anything, email).
					Return(user, nil).
					Once()
				hasherMock.
					On("Check", user.PasswordHash, passwordStr).
					Return(true).
					Once()
				// tokenSvc.GenerateToken возвращает ошибку "token service error"
				tokenSvcMock.
					On("GenerateToken", user.ID, user.Role).
					Return("", errors.New("token service error")).
					Once()
				// Логгер должен залогировать оригинальную ошибку от GenerateToken
				loggerMock.
					On("Errorw", "Login", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "token service error")
					}), "userID", user.ID).
					Return().
					Once()
			case "":
				// Успешный сценарий
				repoMock.
					On("FindByEmail", mock.Anything, email).
					Return(user, nil).
					Once()
				hasherMock.
					On("Check", user.PasswordHash, passwordStr).
					Return(true).
					Once()
				tokenSvcMock.
					On("GenerateToken", user.ID, user.Role).
					Return(fakeToken, nil).
					Once()
			}

			tokenStr, err := svc.Login(ctx, email, passwordStr)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && tokenStr != fakeToken {
				t.Errorf("expected token %q, got %q", fakeToken, tokenStr)
			}
		})
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	validRole := "client"
	invalidRole := "unknown"
	fakeToken := "dummy_jwt_token"

	allowedRoles := map[string]bool{
		"client":    true,
		"moderator": true,
	}

	tests := []struct {
		name           string
		role           string
		allowedRoles   map[string]bool
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "invalid role",
			role:           invalidRole,
			allowedRoles:   allowedRoles,
			expectedErrMsg: "role 'unknown' is not allowed",
		},
		{
			name:           "error in token generate",
			role:           validRole,
			allowedRoles:   allowedRoles,
			simulateError:  "token",
			expectedErrMsg: "dummy login failed",
		},
		{
			name:         "success",
			role:         validRole,
			allowedRoles: allowedRoles,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loggerMock := mockLog.NewLogger(t)
			tokenSvcMock := mockToken.NewTokenService(t)

			svc := &authServiceImp{
				tokenSvc:     tokenSvcMock,
				logger:       loggerMock,
				allowedRoles: tc.allowedRoles,
			}

			// Если роль недопустимая, метод validateRole вернёт ошибку
			if tc.expectedErrMsg != "" && strings.Contains(tc.expectedErrMsg, "not allowed") {
				loggerMock.
					On("Errorw", "DummyLogin", "role", tc.role, "error", mock.Anything).
					Return().
					Once()
				token, err := svc.DummyLogin(ctx, tc.role)
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v (token: %q)", tc.expectedErrMsg, err, token)
				}
				return
			}

			// Для остальных сценариев вызываем tokenSvc.GenerateToken
			switch tc.simulateError {
			case "token":
				tokenSvcMock.
					On("GenerateToken", "dummyID", tc.role).
					Return("", errors.New("token service error")).
					Once()
				loggerMock.
					On("Errorw", "DummyLogin", "role", tc.role, "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "token service error")
					})).
					Return().
					Once()
			case "":
				tokenSvcMock.
					On("GenerateToken", "dummyID", tc.role).
					Return(fakeToken, nil).
					Once()
			}

			token, err := svc.DummyLogin(ctx, tc.role)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && token != fakeToken {
				t.Errorf("expected token %q, got %q", fakeToken, token)
			}
		})
	}
}

func TestAuthService_validateRole(t *testing.T) {
	t.Parallel()

	allowedRoles := map[string]bool{
		"client":    true,
		"moderator": true,
	}

	svc := &authServiceImp{
		allowedRoles: allowedRoles,
	}

	tests := []struct {
		name            string
		role            string
		expectErr       bool
		expectedMessage string
		expectedCode    string
	}{
		{
			name:      "allowed role - client",
			role:      "client",
			expectErr: false,
		},
		{
			name:      "allowed role - moderator",
			role:      "moderator",
			expectErr: false,
		},
		{
			name:            "not allowed role - admin",
			role:            "admin",
			expectErr:       true,
			expectedMessage: "role 'admin' is not allowed",
			expectedCode:    errs.ErrInvalidRoleCode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := svc.validateRole(tt.role)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for role %q, got nil", tt.role)
				}
				if !strings.Contains(err.Error(), tt.expectedMessage) {
					t.Errorf("expected error message to contain %q, got %q", tt.expectedMessage, err.Error())
				}
				var appErr *errs.AppError
				if !errors.As(err, &appErr) {
					t.Errorf("expected error of type *errs.AppError, got %T", err)
				} else if appErr.Code != tt.expectedCode {
					t.Errorf("expected error code %q, got %q", tt.expectedCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect error for role %q, got %v", tt.role, err)
				}
			}
		})
	}
}
