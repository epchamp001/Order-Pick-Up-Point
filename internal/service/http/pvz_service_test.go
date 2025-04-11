package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"order-pick-up-point/internal/errs"
	"order-pick-up-point/internal/models/entity"
	mockRepo "order-pick-up-point/internal/storage/db/mock"
	mockLog "order-pick-up-point/pkg/logger/mock"
	"strings"
	"testing"
	"time"
)

func TestPvzService_CreatePvz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		city           string
		allowedCities  map[string]bool
		repoReturnErr  error
		expectedPvzID  string
		expectedErrMsg string
	}{
		{
			name:           "disallowed city",
			city:           "London",
			allowedCities:  map[string]bool{"moscow": true, "saint petersburg": true, "kazan": true},
			repoReturnErr:  nil,
			expectedPvzID:  "",
			expectedErrMsg: "city 'London' is not allowed",
		},
		{
			name:           "repository error",
			city:           "Moscow",
			allowedCities:  map[string]bool{"moscow": true},
			repoReturnErr:  errors.New("db error"),
			expectedPvzID:  "",
			expectedErrMsg: "db error",
		},
		{
			name:           "success",
			city:           "Moscow",
			allowedCities:  map[string]bool{"moscow": true},
			repoReturnErr:  nil,
			expectedPvzID:  "pvz123",
			expectedErrMsg: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:          repoMock,
				logger:        loggerMock,
				txManager:     txManager,
				allowedCities: tc.allowedCities,
			}

			ctx := context.Background()

			if tc.allowedCities[strings.ToLower(tc.city)] {
				repoMock.
					On("CreatePvz", ctx, mock.MatchedBy(func(pvz entity.Pvz) bool {
						return strings.ToLower(pvz.City) == strings.ToLower(tc.city) &&
							time.Since(pvz.RegistrationDate) < 5*time.Second
					})).
					Return(func(_ context.Context, _ entity.Pvz) string {
						return tc.expectedPvzID
					}, tc.repoReturnErr).
					Once()

				if tc.repoReturnErr != nil {
					loggerMock.
						On("Errorw", "CreatePvz", "error", tc.repoReturnErr, "city", tc.city).
						Once()
				}
			}

			pvzID, err := svc.CreatePvz(ctx, tc.city)

			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got: %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if pvzID != tc.expectedPvzID {
				t.Errorf("expected pvzID %q, got %q", tc.expectedPvzID, pvzID)
			}
		})
	}
}

func TestPvzService_GetPvzsInfo(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now
	pvz := entity.Pvz{ID: "pvz1", City: "Moscow", RegistrationDate: now}
	reception := entity.Reception{ID: "rec1"}
	product := entity.Product{ID: "prod1"}

	tests := []struct {
		name            string
		page            int
		limit           int
		startDate       *time.Time
		endDate         *time.Time
		simulateError   string
		expectedErrMsg  string
		expectedResults []entity.PvzInfo
	}{
		{
			name:           "tx error",
			page:           1,
			limit:          10,
			startDate:      &startDate,
			endDate:        &endDate,
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "error in GetPvzs",
			page:           1,
			limit:          10,
			startDate:      &startDate,
			endDate:        &endDate,
			simulateError:  "pvzs",
			expectedErrMsg: "failed to get pvzs",
		},
		{
			name:           "error in GetReceptionsByPvzIDFiltered",
			page:           1,
			limit:          10,
			startDate:      &startDate,
			endDate:        &endDate,
			simulateError:  "receptions",
			expectedErrMsg: fmt.Sprintf("failed to get receptions for pvz %s", pvz.ID),
		},
		{
			name:           "error in GetProductsByReceptionID",
			page:           1,
			limit:          10,
			startDate:      &startDate,
			endDate:        &endDate,
			simulateError:  "products",
			expectedErrMsg: fmt.Sprintf("failed to get products for reception %s", reception.ID),
		},
		{
			name:          "success",
			page:          1,
			limit:         10,
			startDate:     &startDate,
			endDate:       &endDate,
			simulateError: "",
			expectedResults: []entity.PvzInfo{
				{
					Pvz: pvz,
					Receptions: []entity.ReceptionInfo{
						{
							Reception: reception,
							Products:  []entity.Product{product},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:      repoMock,
				logger:    loggerMock,
				txManager: txManager,
			}
			ctx := context.Background()

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", ctx, pgx.ReadCommitted, pgx.ReadOnly, mock.Anything).
					Return(errors.New("tx error")).
					Once()
				loggerMock.
					On("Errorw", "GetPvzsInfo transaction failed", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					})).
					Return().
					Once()
			} else {
				var callbackErr error
				txManager.
					On("WithTx", ctx, pgx.ReadCommitted, pgx.ReadOnly, mock.Anything).
					Run(func(args mock.Arguments) {
						f := args.Get(3).(func(context.Context) error)
						callbackErr = f(ctx)
					}).
					Return(func(ctx context.Context, iso pgx.TxIsoLevel, access pgx.TxAccessMode, f func(context.Context) error) error {
						return callbackErr
					}).
					Once()

				switch tc.simulateError {
				case "pvzs":
					repoMock.
						On("GetPvzs", ctx, tc.page, tc.limit).
						Return(nil, errors.New("pvzs error")).
						Once()
					loggerMock.
						On("Errorw", "get PVZs", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), "pvzs error")
						}), "page", tc.page, "limit", tc.limit).
						Return().
						Once()
				case "receptions":
					repoMock.
						On("GetPvzs", ctx, tc.page, tc.limit).
						Return([]entity.Pvz{pvz}, nil).
						Once()
					repoMock.
						On("GetReceptionsByPvzIDFiltered", ctx, pvz.ID, tc.startDate, tc.endDate).
						Return(nil, errors.New("receptions error")).
						Once()
					loggerMock.
						On("Errorw", "get receptions", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), "receptions error")
						}), "pvzID", pvz.ID).
						Return().
						Once()
				case "products":
					repoMock.
						On("GetPvzs", ctx, tc.page, tc.limit).
						Return([]entity.Pvz{pvz}, nil).
						Once()
					repoMock.
						On("GetReceptionsByPvzIDFiltered", ctx, pvz.ID, tc.startDate, tc.endDate).
						Return([]entity.Reception{reception}, nil).
						Once()
					repoMock.
						On("GetProductsByReceptionID", ctx, reception.ID).
						Return(nil, errors.New("products error")).
						Once()
					loggerMock.
						On("Errorw", "get products", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), "products error")
						}), "receptionID", reception.ID).
						Return().
						Once()
				case "":
					repoMock.
						On("GetPvzs", ctx, tc.page, tc.limit).
						Return([]entity.Pvz{pvz}, nil).
						Once()
					repoMock.
						On("GetReceptionsByPvzIDFiltered", ctx, pvz.ID, tc.startDate, tc.endDate).
						Return([]entity.Reception{reception}, nil).
						Once()
					repoMock.
						On("GetProductsByReceptionID", ctx, reception.ID).
						Return([]entity.Product{product}, nil).
						Once()
				}

				if tc.expectedErrMsg != "" {
					loggerMock.
						On("Errorw", "GetPvzsInfo transaction failed", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						})).
						Return().
						Once()
				}
			}

			result, err := svc.GetPvzsInfo(ctx, tc.page, tc.limit, tc.startDate, tc.endDate)

			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tc.expectedErrMsg == "" {
				if len(result) != len(tc.expectedResults) {
					t.Fatalf("expected %d pvzInfo, got %d", len(tc.expectedResults), len(result))
				}
				for i, exp := range tc.expectedResults {
					res := result[i]
					if res.Pvz.ID != exp.Pvz.ID {
						t.Errorf("expected pvz ID %q, got %q", exp.Pvz.ID, res.Pvz.ID)
					}
					if len(res.Receptions) != len(exp.Receptions) {
						t.Errorf("expected %d receptions, got %d", len(exp.Receptions), len(res.Receptions))
					}
					if len(exp.Receptions) > 0 {
						if res.Receptions[0].Reception.ID != exp.Receptions[0].Reception.ID {
							t.Errorf("expected reception ID %q, got %q", exp.Receptions[0].Reception.ID, res.Receptions[0].Reception.ID)
						}
						if len(res.Receptions[0].Products) != len(exp.Receptions[0].Products) {
							t.Errorf("expected %d products, got %d", len(exp.Receptions[0].Products), len(res.Receptions[0].Products))
						}
						if len(exp.Receptions[0].Products) > 0 && res.Receptions[0].Products[0].ID != exp.Receptions[0].Products[0].ID {
							t.Errorf("expected product ID %q, got %q", exp.Receptions[0].Products[0].ID, res.Receptions[0].Products[0].ID)
						}
					}
				}
			}
		})
	}
}

func TestPvzService_CreateReception(t *testing.T) {
	t.Parallel()

	now := time.Now()
	pvzID := "pvz123"
	fakeReceptionID := "rec456"

	notFoundErr := &errs.AppError{Code: errs.ErrNoOpenReception, Message: "open reception not found"}

	tests := []struct {
		name           string
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "tx error",
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "open reception exists",
			simulateError:  "exists",
			expectedErrMsg: "open reception already exists",
		},
		{
			name:           "error in find open reception (non not found error)",
			simulateError:  "non_not_found_error",
			expectedErrMsg: "find error",
		},
		{
			name:           "error in CreateReception",
			simulateError:  "create",
			expectedErrMsg: "create error",
		},
		{
			name:          "success",
			simulateError: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:      repoMock,
				logger:    loggerMock,
				txManager: txManager,
			}
			ctx := context.Background()

			var callbackErr error

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Return(errors.New("tx error")).
					Once()
				loggerMock.
					On("Errorw", "CreateReception", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					}), "pvzID", pvzID).
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
				case "exists":
					// Симулируем, что открытая приёмка уже существует
					existing := &entity.Reception{ID: "dummy", PvzID: pvzID, DateTime: now, Status: "in_progress"}
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(existing, nil).
						Once()
				case "non_not_found_error":
					// Возвращаем ошибку "find error", которая не содержит "not found"
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, errors.New("find error")).
						Once()
				case "create":
					// Сценарий "create": открытая приёмка не найдена,
					// затем при создании возникает ошибка
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, notFoundErr).
						Once()
					repoMock.
						On("CreateReception", mock.Anything, mock.MatchedBy(func(r entity.Reception) bool {
							return r.PvzID == pvzID && r.Status == "in_progress" && r.DateTime.Equal(now)
						})).
						Return("", errors.New("create error")).
						Once()
				case "":
					// Успешный сценарий: открытая приёмка не найдена, создание проходит успешно
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, notFoundErr).
						Once()
					repoMock.
						On("CreateReception", mock.Anything, mock.MatchedBy(func(r entity.Reception) bool {
							return r.PvzID == pvzID && r.Status == "in_progress" && r.DateTime.Equal(now)
						})).
						Return(fakeReceptionID, nil).
						Once()
				}

				if tc.expectedErrMsg != "" {
					loggerMock.
						On("Errorw", "CreateReception", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						}), "pvzID", pvzID).
						Return().
						Once()
				}
			}

			result, err := svc.CreateReception(ctx, pvzID, now)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && result != fakeReceptionID {
				t.Errorf("expected receptionID %q, got %q", fakeReceptionID, result)
			}
		})
	}
}

func TestPvzService_AddProduct(t *testing.T) {
	t.Parallel()

	// Опорные данные
	ctx := context.Background()
	pvzID := "pvz123"
	validProductType := "electronics"
	invalidProductType := "toys"
	fakeProductID := "prod789"
	now := time.Now()

	// Для успешных сценариев разрешим только "electronics"
	allowedProductTypes := map[string]bool{
		"electronics": true,
	}

	// Подготовленная открытая приёмка (для успешного сценария)
	openReception := &entity.Reception{ID: "rec001", PvzID: pvzID, DateTime: now, Status: "in_progress"}

	tests := []struct {
		name           string
		productType    string
		allowedTypes   map[string]bool
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "invalid product type",
			productType:    invalidProductType,
			allowedTypes:   allowedProductTypes,
			expectedErrMsg: "product type 'toys' is not allowed",
		},
		{
			name:           "tx error",
			productType:    validProductType,
			allowedTypes:   allowedProductTypes,
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "error in FindOpenReception",
			productType:    validProductType,
			allowedTypes:   allowedProductTypes,
			simulateError:  "find",
			expectedErrMsg: "find error",
		},
		{
			name:           "no open reception",
			productType:    validProductType,
			allowedTypes:   allowedProductTypes,
			simulateError:  "no_reception",
			expectedErrMsg: "no open reception found for this PVZ",
		},
		{
			name:           "error in CreateProduct",
			productType:    validProductType,
			allowedTypes:   allowedProductTypes,
			simulateError:  "create",
			expectedErrMsg: "create error",
		},
		{
			name:          "success",
			productType:   validProductType,
			allowedTypes:  allowedProductTypes,
			simulateError: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:                repoMock,
				logger:              loggerMock,
				txManager:           txManager,
				allowedProductTypes: tc.allowedTypes,
			}

			var callbackErr error

			if tc.expectedErrMsg != "" && strings.Contains(tc.expectedErrMsg, "not allowed") {
				result, err := svc.AddProduct(ctx, pvzID, tc.productType)
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v (result: %q)", tc.expectedErrMsg, err, result)
				}
				return
			}

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Return(errors.New("tx error")).
					Once()
				loggerMock.
					On("Errorw", "AddProduct", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					}), "pvzID", pvzID, "productType", tc.productType).
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
				case "find":
					// Ошибка при поиске открытой приёмки
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, errors.New("find error")).
						Once()
				case "no_reception":
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, nil).
						Once()
				case "create":
					// Сначала поиск проходит успешно и возвращает открытую приёмку,
					// затем при вызове CreateProduct возникает ошибка
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(openReception, nil).
						Once()
					repoMock.
						On("CreateProduct", mock.Anything, mock.MatchedBy(func(p entity.Product) bool {
							return p.ReceptionID == openReception.ID && strings.ToLower(p.Type) == strings.ToLower(tc.productType)
						})).
						Return("", errors.New("create error")).
						Once()
				case "":
					// Успешный сценарий: поиск открытой приёмки проходит успешно, создается товар
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(openReception, nil).
						Once()
					repoMock.
						On("CreateProduct", mock.Anything, mock.MatchedBy(func(p entity.Product) bool {
							return p.ReceptionID == openReception.ID && strings.ToLower(p.Type) == strings.ToLower(tc.productType)
						})).
						Return(fakeProductID, nil).
						Once()
				}

				if tc.expectedErrMsg != "" {
					loggerMock.
						On("Errorw", "AddProduct", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						}), "pvzID", pvzID, "productType", tc.productType).
						Return().
						Once()
				}
			}

			result, err := svc.AddProduct(ctx, pvzID, tc.productType)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && result != fakeProductID {
				t.Errorf("expected productID %q, got %q", fakeProductID, result)
			}
		})
	}
}

func TestPvzService_DeleteLastProduct(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pvzID := "pvz123"

	fakeProduct := &entity.Product{ID: "prod001"}

	tests := []struct {
		name           string
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "tx error",
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "error in FindLastProductInReception",
			simulateError:  "find",
			expectedErrMsg: "find error",
		},
		{
			name:           "no product found",
			simulateError:  "no_product",
			expectedErrMsg: "no product found to delete",
		},
		{
			name:           "error in DeleteProduct",
			simulateError:  "delete",
			expectedErrMsg: "delete error",
		},
		{
			name:          "success",
			simulateError: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:      repoMock,
				logger:    loggerMock,
				txManager: txManager,
			}

			var callbackErr error

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Return(errors.New("tx error")).
					Once()
				loggerMock.
					On("Errorw", "DeleteLastProduct", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					}), "pvzID", pvzID).
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
				case "find":
					repoMock.
						On("FindLastProductInReception", mock.Anything, pvzID).
						Return(nil, errors.New("find error")).
						Once()
				case "no_product":
					repoMock.
						On("FindLastProductInReception", mock.Anything, pvzID).
						Return(nil, nil).
						Once()
				case "delete":
					repoMock.
						On("FindLastProductInReception", mock.Anything, pvzID).
						Return(fakeProduct, nil).
						Once()
					repoMock.
						On("DeleteProduct", mock.Anything, fakeProduct.ID).
						Return(errors.New("delete error")).
						Once()
				case "":
					// Успешный сценарий: продукт найден, удаление проходит успешно
					repoMock.
						On("FindLastProductInReception", mock.Anything, pvzID).
						Return(fakeProduct, nil).
						Once()
					repoMock.
						On("DeleteProduct", mock.Anything, fakeProduct.ID).
						Return(nil).
						Once()
				}

				if tc.expectedErrMsg != "" {
					loggerMock.
						On("Errorw", "DeleteLastProduct", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						}), "pvzID", pvzID).
						Return().
						Once()
				}
			}

			err := svc.DeleteLastProduct(ctx, pvzID)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPvzService_CloseReception(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pvzID := "pvz123"
	openReception := &entity.Reception{ID: "rec001", PvzID: pvzID, Status: "in_progress"}

	tests := []struct {
		name           string
		simulateError  string
		expectedErrMsg string
	}{
		{
			name:           "tx error",
			simulateError:  "tx",
			expectedErrMsg: "tx error",
		},
		{
			name:           "error in FindOpenReception",
			simulateError:  "find",
			expectedErrMsg: "find error",
		},
		{
			name:           "no open reception found",
			simulateError:  "not_found",
			expectedErrMsg: "no open reception found for this PVZ",
		},
		{
			name:           "error in UpdateReceptionStatus",
			simulateError:  "update",
			expectedErrMsg: "update error",
		},
		{
			name:          "success",
			simulateError: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockRepo.NewRepository(t)
			loggerMock := mockLog.NewLogger(t)
			txManager := mockRepo.NewTxManager(t)

			svc := &pvzServiceImp{
				repo:      repoMock,
				logger:    loggerMock,
				txManager: txManager,
			}

			var callbackErr error

			if tc.simulateError == "tx" {
				txManager.
					On("WithTx", mock.Anything, pgx.ReadCommitted, pgx.ReadWrite, mock.Anything).
					Return(errors.New("tx error")).
					Once()

				loggerMock.
					On("Errorw", "CloseReception", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), "tx error")
					}), "pvzID", pvzID).
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
				case "find":
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, errors.New("find error")).
						Once()
				case "not_found":
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(nil, nil).
						Once()
				case "update":
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(openReception, nil).
						Once()
					repoMock.
						On("UpdateReceptionStatus", mock.Anything, openReception.ID, "close").
						Return(errors.New("update error")).
						Once()
				case "":
					// Успешный сценарий: открытая приёмка найдена, обновление статуса проходит успешно
					repoMock.
						On("FindOpenReceptionByPvzID", mock.Anything, pvzID).
						Return(openReception, nil).
						Once()
					repoMock.
						On("UpdateReceptionStatus", mock.Anything, openReception.ID, "close").
						Return(nil).
						Once()
				}

				if tc.expectedErrMsg != "" {
					loggerMock.
						On("Errorw", "CloseReception", "error", mock.MatchedBy(func(err error) bool {
							return strings.Contains(err.Error(), tc.expectedErrMsg)
						}), "pvzID", pvzID).
						Return().
						Once()
				}
			}

			recID, err := svc.CloseReception(ctx, pvzID)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectedErrMsg == "" && recID != openReception.ID {
				t.Errorf("expected reception ID %q, got %q", openReception.ID, recID)
			}
		})
	}
}
