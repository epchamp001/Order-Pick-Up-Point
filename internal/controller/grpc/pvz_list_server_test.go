package grpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"order-pick-up-point/api/pb"
	"order-pick-up-point/internal/models/entity"
	mockPvzSvc "order-pick-up-point/internal/service/grpc/mock"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPvzServer_GetPVZList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Список PVZ для успешного сценария
	dummyList := []entity.Pvz{
		{ID: "1", City: "Moscow", RegistrationDate: time.Date(2023, 10, 5, 15, 0, 0, 0, time.UTC)},
		{ID: "2", City: "Kazan", RegistrationDate: time.Date(2023, 10, 6, 15, 0, 0, 0, time.UTC)},
	}

	tests := []struct {
		name                 string
		simulateError        bool
		errToReturn          error
		expectedErrSubstring string
		expectedList         []entity.Pvz
	}{
		{
			name:                 "error in GetPVZList",
			simulateError:        true,
			errToReturn:          errors.New("database error"),
			expectedErrSubstring: "failed to get PVZ list",
			expectedList:         nil,
		},
		{
			name:                 "success",
			simulateError:        false,
			errToReturn:          nil,
			expectedErrSubstring: "",
			expectedList:         dummyList,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := mockPvzSvc.NewPvzService(t)
			mockSvc.
				On("GetPVZList", mock.Anything).
				Return(tc.expectedList, tc.errToReturn).
				Once()

			server := NewPvzServer(mockSvc)

			req := &pb.GetPVZListRequest{}

			resp, err := server.GetPVZList(ctx, req)
			if tc.simulateError {
				// В случае ошибки, контроллер должен вернуть обёрнутую ошибку с сообщением "failed to get PVZ list"
				if err == nil || !strings.Contains(err.Error(), tc.expectedErrSubstring) {
					t.Errorf("expected error containing %q, got %v", tc.expectedErrSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Проверяем, что количество элементов соответствует
				if len(resp.Pvzs) != len(tc.expectedList) {
					t.Errorf("expected %d PVZ entries, got %d", len(tc.expectedList), len(resp.Pvzs))
				}

				// Проверяем поля каждого объекта
				for i, pvz := range tc.expectedList {
					pbPvz := resp.Pvzs[i]
					if pbPvz.Id != pvz.ID {
						t.Errorf("expected pvz ID %q, got %q", pvz.ID, pbPvz.Id)
					}
					if pbPvz.City != pvz.City {
						t.Errorf("expected city %q, got %q", pvz.City, pbPvz.City)
					}
					expectedTs := timestamppb.New(pvz.RegistrationDate)
					if !reflect.DeepEqual(pbPvz.RegistrationDate, expectedTs) {
						t.Errorf("expected RegistrationDate %+v, got %+v", expectedTs, pbPvz.RegistrationDate)
					}
				}
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
