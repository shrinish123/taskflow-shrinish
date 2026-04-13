package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/repository"
)

type AppError struct {
	Status  int
	Message string
	Err     error
}

func (e *AppError) Error() string       { return e.Message }
func (e *AppError) Unwrap() error        { return e.Err }
func (e *AppError) HTTPStatus() int      { return e.Status }
func (e *AppError) HTTPMessage() string  { return e.Message }

var (
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

func NotFoundErr(resource string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Message: resource + " not found",
		Err:     ErrNotFound,
	}
}

func ForbiddenErr(message string) *AppError {
	return &AppError{
		Status:  http.StatusForbidden,
		Message: message,
		Err:     ErrForbidden,
	}
}

func EmailTakenErr() *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Message: "email is already taken",
		Err:     ErrEmailTaken,
	}
}

func InvalidCredentialsErr() *AppError {
	return &AppError{
		Status:  http.StatusUnauthorized,
		Message: "invalid email or password",
		Err:     ErrInvalidCredentials,
	}
}

func InternalErr(err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
		Err:     fmt.Errorf("internal: %w", err),
	}
}

func mapRepoErr(err error, resource string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return NotFoundErr(resource)
	}
	if errors.Is(err, repository.ErrDuplicateEmail) {
		return EmailTakenErr()
	}
	return InternalErr(err)
}
