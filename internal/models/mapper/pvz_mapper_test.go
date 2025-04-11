package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
	"reflect"
	"testing"
	"time"
)

func TestPvzEntityToDTO(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name     string
		input    entity.Pvz
		expected dto.PvzDTO
	}{
		{
			name: "normal conversion",
			input: entity.Pvz{
				ID:               "1",
				RegistrationDate: now,
				City:             "Moscow",
			},
			expected: dto.PvzDTO{
				Id:               "1",
				RegistrationDate: now,
				City:             "Moscow",
			},
		},
		{
			name: "empty PVZ",
			input: entity.Pvz{
				ID:               "",
				RegistrationDate: time.Time{},
				City:             "",
			},
			expected: dto.PvzDTO{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := PvzEntityToDTO(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, result)
			}
		})
	}
}

func TestPvzDTOToEntity(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name     string
		input    dto.PvzDTO
		expected entity.Pvz
	}{
		{
			name: "normal conversion",
			input: dto.PvzDTO{
				Id:               "1",
				RegistrationDate: now,
				City:             "Moscow",
			},
			expected: entity.Pvz{
				ID:               "1",
				RegistrationDate: now,
				City:             "Moscow",
			},
		},
		{
			name:     "empty DTO",
			input:    dto.PvzDTO{},
			expected: entity.Pvz{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := PvzDTOToEntity(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, result)
			}
		})
	}
}

func TestPvzInfoEntityToResponse(t *testing.T) {
	t.Parallel()

	now := time.Now()

	pvzEntity := entity.Pvz{
		ID:               "1",
		RegistrationDate: now,
		City:             "Moscow",
	}

	receptionEntity := entity.Reception{
		ID: "r1",
	}

	productEntity := entity.Product{
		ID:          "p1",
		DateTime:    now,
		Type:        "electronics",
		ReceptionID: "r1",
	}

	recInfo := entity.ReceptionInfo{
		Reception: receptionEntity,
		Products:  []entity.Product{productEntity},
	}

	infoEntity := entity.PvzInfo{
		Pvz:        pvzEntity,
		Receptions: []entity.ReceptionInfo{recInfo},
	}

	expectedPvzDTO := dto.PvzDTO{
		Id:               "1",
		RegistrationDate: now,
		City:             "Moscow",
	}

	expectedRecDTO := dto.ReceptionDTO{
		Id: "r1",
	}
	expectedProdDTO := dto.ProductDTO{
		Id:          "p1",
		DateTime:    now,
		Type:        "electronics",
		ReceptionId: "r1",
	}

	expectedResponse := dto.PvzGet200ResponseInner{
		Pvz: &expectedPvzDTO,
		Receptions: []dto.PvzGet200ResponseInnerReceptionsInner{
			{
				Reception: &expectedRecDTO,
				Products:  []dto.ProductDTO{expectedProdDTO},
			},
		},
	}

	result := PvzInfoEntityToResponse(infoEntity)
	if !reflect.DeepEqual(result, expectedResponse) {
		t.Errorf("expected %+v, got %+v", expectedResponse, result)
	}
}
