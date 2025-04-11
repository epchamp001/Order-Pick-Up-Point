package validator

import (
	"order-pick-up-point/internal/errs"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^.+@[A-Za-z0-9]+\.[A-Za-z0-9]+$`)

const forbiddenChars = `()<>[]:;,"`

func ValidateEmail(email string) error {
	if email == "" {
		return errs.New(errs.ErrInvalidEmail, "email must not be empty")
	}

	if strings.ContainsAny(email, forbiddenChars) {
		return errs.New(errs.ErrInvalidEmail, "email contains forbidden characters")
	}

	if !emailRegex.MatchString(email) {
		return errs.New(errs.ErrInvalidEmail, "invalid email format")
	}
	return nil
}
