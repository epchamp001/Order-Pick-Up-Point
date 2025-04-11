package validator

import "order-pick-up-point/internal/errs"

func ValidatePassword(password string) error {
	if len(password) < 4 {
		return errs.New(errs.ErrWeakPassword, "password must be at least 4 characters long")
	}

	return nil
}
