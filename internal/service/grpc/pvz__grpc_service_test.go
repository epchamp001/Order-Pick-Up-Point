package grpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"order-pick-up-point/internal/models/entity"
	mockPvzRepo "order-pick-up-point/internal/storage/db/mock"
	mockLog "order-pick-up-point/pkg/logger/mock"
	"strings"
	"testing"
	"time"
)

func TestPvzService_GetPVZList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Данные для успешного сценария
	dummyList := []entity.Pvz{
		{ID: "1", City: "Moscow", RegistrationDate: time.Now()},
		{ID: "2", City: "Kazan", RegistrationDate: time.Now()},
	}

	tests := []struct {
		name           string
		simulateError  bool
		errToReturn    error
		expectedErrMsg string
		expectedList   []entity.Pvz
	}{
		{
			name:           "error in GetListOfPvzs",
			simulateError:  true,
			errToReturn:    errors.New("db error"),
			expectedErrMsg: "db error",
			expectedList:   nil,
		},
		{
			name:           "success",
			simulateError:  false,
			errToReturn:    nil,
			expectedErrMsg: "",
			expectedList:   dummyList,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repoMock := mockPvzRepo.NewPvzRepository(t)
			loggerMock := mockLog.NewLogger(t)

			svc := &pvzServiceImp{
				repo:   repoMock,
				logger: loggerMock,
			}

			if tc.simulateError {
				// Если ожидается ошибка, репозиторий возвращает ошибку
				repoMock.
					On("GetListOfPvzs", mock.Anything).
					Return(nil, tc.errToReturn).
					Once()
				loggerMock.
					On("Errorw", "get PVZ list", "error", mock.MatchedBy(func(err error) bool {
						return strings.Contains(err.Error(), tc.expectedErrMsg)
					})).
					Return().
					Once()
			} else {
				repoMock.
					On("GetListOfPvzs", mock.Anything).
					Return(tc.expectedList, nil).
					Once()
			}

			pvzs, err := svc.GetPVZList(ctx)
			if tc.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(pvzs) != len(tc.expectedList) {
					t.Errorf("expected %d pvzs, got %d", len(tc.expectedList), len(pvzs))
				}
			}
		})
	}
}
