package validator

import (
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		password   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "password too short",
			password:   "abc",
			wantErr:    true,
			wantErrMsg: "password must be at least 4 characters long",
		},
		{
			name:     "valid password with exactly 4 characters",
			password: "abcd",
			wantErr:  false,
		},
		{
			name:     "long valid password",
			password: "thisIsASecurePassword",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePassword(tc.password)
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
