package errors

import (
	"errors"
	"net/http"
)

// Sentinel domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternal           = errors.New("internal server error")
)

// AppError is a structured application error
type AppError struct {
	HTTPCode int
	Message  string
	Err      error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New constructs an AppError
func New(httpCode int, message string, err error) *AppError {
	return &AppError{HTTPCode: httpCode, Message: message, Err: err}
}

// Map maps domain sentinel errors to AppError
func Map(err error) *AppError {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return New(http.StatusNotFound, err.Error(), err)
	case errors.Is(err, ErrEmailAlreadyExists):
		return New(http.StatusConflict, err.Error(), err)
	case errors.Is(err, ErrInvalidCredentials):
		return New(http.StatusUnauthorized, err.Error(), err)
	case errors.Is(err, ErrInvalidInput):
		return New(http.StatusBadRequest, err.Error(), err)
	case errors.Is(err, ErrUnauthorized):
		return New(http.StatusUnauthorized, err.Error(), err)
	case errors.Is(err, ErrForbidden):
		return New(http.StatusForbidden, err.Error(), err)
	default:
		return New(http.StatusInternalServerError, "internal server error", err)
	}
}
