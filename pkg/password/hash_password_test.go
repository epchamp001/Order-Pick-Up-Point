package password

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestNewBCryptHasher(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputCost    int
		expectedCost int
	}{
		{
			name:         "Negative cost uses default",
			inputCost:    -1,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "Zero cost uses default",
			inputCost:    0,
			expectedCost: bcrypt.DefaultCost,
		},
		{
			name:         "Positive cost uses provided value",
			inputCost:    12,
			expectedCost: 12,
		},
	}

	for _, tc := range tests {
		tc := tc // захватываем переменную
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hasher := NewBCryptHasher(tc.inputCost)
			bHasher, ok := hasher.(*BCryptHasher)
			if !ok {
				t.Fatalf("Expected *BCryptHasher, got %T", hasher)
			}
			if bHasher.cost != tc.expectedCost {
				t.Errorf("Expected cost %d, got %d", tc.expectedCost, bHasher.cost)
			}
		})
	}
}

func TestBCryptHasher_Hash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		password    string
		cost        int
		expectError bool
	}{
		{
			name:        "Valid password and default cost",
			password:    "password123",
			cost:        bcrypt.DefaultCost,
			expectError: false,
		},
		{
			name:        "Valid password and custom cost",
			password:    "anotherPassword!",
			cost:        10,
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			cost:        bcrypt.DefaultCost,
			expectError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hasher := NewBCryptHasher(tc.cost)
			hash, err := hasher.Hash(tc.password)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			if hash != "" {
				if !hasher.Check(hash, tc.password) {
					t.Errorf("Hashed password did not validate with correct password")
				}
			}
		})
	}
}

func TestBCryptHasher_Check(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		password   string
		modifyHash bool
	}{
		{
			name:       "Valid password check",
			password:   "securePassword!",
			modifyHash: false,
		},
		{
			name:       "Invalid password check",
			password:   "anotherSecurePassword!",
			modifyHash: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			hasher := NewBCryptHasher(bcrypt.DefaultCost)
			hash, err := hasher.Hash("securePassword!")
			if err != nil {
				t.Fatalf("Hash generation failed: %v", err)
			}

			if tc.modifyHash {
				// Меняем последний символ, что гарантированно сделает сравнение неуспешным.
				if len(hash) > 0 {
					lastChar := hash[len(hash)-1]
					var newChar byte
					if lastChar == 'a' {
						newChar = 'b'
					} else {
						newChar = 'a'
					}
					hash = hash[:len(hash)-1] + string(newChar)
				}
			}

			ok := hasher.Check(hash, tc.password)
			if !tc.modifyHash && !ok {
				t.Errorf("Expected password validation to succeed")
			}
			if tc.modifyHash && ok {
				t.Errorf("Expected password validation to fail with modified hash")
			}
		})
	}
}
