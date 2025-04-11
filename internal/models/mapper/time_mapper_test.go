package mapper

import (
	"reflect"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	t.Parallel()

	validTimeStr := "2023-10-05T14:48:00Z"
	expectedTime, err := time.Parse(timeLayout, validTimeStr)
	if err != nil {
		t.Fatalf("failed to parse validTimeStr: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		expectedTime *time.Time
		expectingErr bool
	}{
		{
			name:         "empty string returns nil",
			input:        "",
			expectedTime: nil,
			expectingErr: false,
		},
		{
			name:         "valid time string",
			input:        validTimeStr,
			expectedTime: &expectedTime,
			expectingErr: false,
		},
		{
			name:         "invalid time string",
			input:        "invalid-time",
			expectedTime: nil,
			expectingErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := ParseTime(tc.input)
			if tc.expectingErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if !reflect.DeepEqual(res, tc.expectedTime) {
					t.Errorf("expected %v, got %v", tc.expectedTime, res)
				}
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2023, time.October, 5, 14, 48, 0, 0, time.UTC)
	formatted := testTime.Format(timeLayout)

	tests := []struct {
		name        string
		input       *time.Time
		expectedStr string
	}{
		{
			name:        "nil time returns empty string",
			input:       nil,
			expectedStr: "",
		},
		{
			name:        "valid time returns formatted string",
			input:       &testTime,
			expectedStr: formatted,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := FormatTime(tc.input)
			if res != tc.expectedStr {
				t.Errorf("expected %q, got %q", tc.expectedStr, res)
			}
		})
	}
}
