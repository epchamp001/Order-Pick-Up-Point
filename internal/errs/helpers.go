package errs

import (
	"errors"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrNotFoundCode
	}
	return false
}

func IsOpenReceptionNotFound(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrNotFoundCode || appErr.Code == ErrNoOpenReception
	}
	return false
}
