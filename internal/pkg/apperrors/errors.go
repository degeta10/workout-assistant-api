package apperrors

import (
	"errors"
	"net/http"
)

// AppError is a typed application error that carries HTTP semantics.
type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(statusCode int, code, message string, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

func Validation(err error) *AppError {
	return New(http.StatusUnprocessableEntity, "validation_error", "The given data was invalid.", err)
}

func BadRequest(message string, err error) *AppError {
	return New(http.StatusBadRequest, "bad_request", message, err)
}

func Unauthorized(message string, err error) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", message, err)
}

func Forbidden(message string, err error) *AppError {
	return New(http.StatusForbidden, "forbidden", message, err)
}

func NotFound(message string, err error) *AppError {
	return New(http.StatusNotFound, "not_found", message, err)
}

func Conflict(message string, err error) *AppError {
	return New(http.StatusConflict, "conflict", message, err)
}

func Internal(message string, err error) *AppError {
	return New(http.StatusInternalServerError, "internal_error", message, err)
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
