package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
	"reflect"
	"testing"
	"time"
)

func Test_ProductEntityToDTO(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    entity.Product
		expected dto.ProductDTO
	}{
		{
			name: "all fields set",
			input: entity.Product{
				ID:          "prod1",
				DateTime:    now,
				Type:        "electronics",
				ReceptionID: "rec1",
			},
			expected: dto.ProductDTO{
				Id:          "prod1",
				DateTime:    now,
				Type:        "electronics",
				ReceptionId: "rec1",
			},
		},
		{
			name: "empty fields",
			input: entity.Product{
				ID:          "",
				DateTime:    time.Time{},
				Type:        "",
				ReceptionID: "",
			},
			expected: dto.ProductDTO{
				Id:          "",
				DateTime:    time.Time{},
				Type:        "",
				ReceptionId: "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dtoResult := ProductEntityToDTO(tc.input)
			if !reflect.DeepEqual(dtoResult, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, dtoResult)
			}
		})
	}
}

func Test_ProductDTOToEntity(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    dto.ProductDTO
		expected entity.Product
	}{
		{
			name: "all fields set",
			input: dto.ProductDTO{
				Id:          "prod1",
				DateTime:    now,
				Type:        "electronics",
				ReceptionId: "rec1",
			},
			expected: entity.Product{
				ID:          "prod1",
				DateTime:    now,
				Type:        "electronics",
				ReceptionID: "rec1",
			},
		},
		{
			name: "empty fields",
			input: dto.ProductDTO{
				Id:          "",
				DateTime:    time.Time{},
				Type:        "",
				ReceptionId: "",
			},
			expected: entity.Product{
				ID:          "",
				DateTime:    time.Time{},
				Type:        "",
				ReceptionID: "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entityResult := ProductDTOToEntity(tc.input)
			if !reflect.DeepEqual(entityResult, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, entityResult)
			}
		})
	}
}
