package errs

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name: "AppError with ErrNotFoundCode returns true",
			err: &AppError{
				Code:    ErrNotFoundCode,
				Message: "not found",
			},
			expected: true,
		},
		{
			name: "AppError with different code returns false",
			err: &AppError{
				Code:    "SOME_OTHER_CODE",
				Message: "error",
			},
			expected: false,
		},
		{
			name:     "non-AppError returns false",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsNotFound(tc.err)
			if got != tc.expected {
				t.Errorf("IsNotFound(%v)= %v; expected %v", tc.err, got, tc.expected)
			}
		})
	}
}

func TestIsOpenReceptionNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error returns false",
			err:      nil,
			expected: false,
		},
		{
			name: "AppError with ErrNotFoundCode returns true",
			err: &AppError{
				Code:    ErrNotFoundCode,
				Message: "not found",
			},
			expected: true,
		},
		{
			name: "AppError with ErrNoOpenReception returns true",
			err: &AppError{
				Code:    ErrNoOpenReception,
				Message: "no open reception",
			},
			expected: true,
		},
		{
			name: "AppError with different code returns false",
			err: &AppError{
				Code:    "SOME_OTHER_CODE",
				Message: "error",
			},
			expected: false,
		},
		{
			name:     "non-AppError returns false",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := IsOpenReceptionNotFound(tc.err)
			if got != tc.expected {
				t.Errorf("IsOpenReceptionNotFound(%v)= %v; expected %v", tc.err, got, tc.expected)
			}
		})
	}
}
