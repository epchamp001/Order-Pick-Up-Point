package validator

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		email      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "empty email",
			email:      "",
			wantErr:    true,
			wantErrMsg: "email must not be empty",
		},
		{
			name:       "forbidden characters",
			email:      "user,example@example.com",
			wantErr:    true,
			wantErrMsg: "email contains forbidden characters",
		},
		{
			name:       "invalid email format",
			email:      "not-an-email",
			wantErr:    true,
			wantErrMsg: "invalid email format",
		},
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateEmail(tc.email)
			if tc.wantErr && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("expected error message to contain %q, got: %v", tc.wantErrMsg, err.Error())
			}
		})
	}
}
