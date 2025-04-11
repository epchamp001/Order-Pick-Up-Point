package mapper

import (
	"order-pick-up-point/internal/models/dto"
	"order-pick-up-point/internal/models/entity"
	"reflect"
	"testing"
	"time"
)

func Test_ReceptionEntityToDTO(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    entity.Reception
		expected dto.ReceptionDTO
	}{
		{
			name: "all fields set",
			input: entity.Reception{
				ID:       "r1",
				DateTime: now,
				PvzID:    "pvz1",
				Status:   "in_progress",
			},
			expected: dto.ReceptionDTO{
				Id:       "r1",
				DateTime: now,
				PvzId:    "pvz1",
				Status:   "in_progress",
			},
		},
		{
			name: "empty fields",
			input: entity.Reception{
				ID:       "",
				DateTime: time.Time{},
				PvzID:    "",
				Status:   "",
			},
			expected: dto.ReceptionDTO{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ReceptionEntityToDTO(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, got)
			}
		})
	}
}

func Test_ReceptionDTOToEntity(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		input    dto.ReceptionDTO
		expected entity.Reception
	}{
		{
			name: "all fields set",
			input: dto.ReceptionDTO{
				Id:       "r1",
				DateTime: now,
				PvzId:    "pvz1",
				Status:   "in_progress",
			},
			expected: entity.Reception{
				ID:       "r1",
				DateTime: now,
				PvzID:    "pvz1",
				Status:   "in_progress",
			},
		},
		{
			name: "empty fields",
			input: dto.ReceptionDTO{
				Id:       "",
				DateTime: time.Time{},
				PvzId:    "",
				Status:   "",
			},
			expected: entity.Reception{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ReceptionDTOToEntity(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, got)
			}
		})
	}
}
