package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeBadRequest   ErrorCode = "BAD_REQUEST"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeRateLimited  ErrorCode = "RATE_LIMITED"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Details []FieldError
}

func (e *ValidationError) Error() string { return "validasi input gagal" }

type APIError struct {
	Status  int
	Code    ErrorCode
	Message string
}

func (e *APIError) Error() string { return e.Message }

func NewAPIError(status int, code ErrorCode, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

var ErrNotFound = NewAPIError(404, CodeNotFound, "Resource tidak ditemukan")

var ErrInvalidCredentials = NewAPIError(401, CodeUnauthorized, "Email atau password salah")

var ErrSessionRevoked = NewAPIError(401, CodeUnauthorized, "Sesi tidak valid, silakan login kembali")

func BadRequest(message string) *APIError {
	return NewAPIError(400, CodeBadRequest, message)
}

func Unauthorized(message string) *APIError {
	return NewAPIError(401, CodeUnauthorized, message)
}

func Forbidden(message string) *APIError {
	return NewAPIError(403, CodeForbidden, message)
}

func Conflict(message string) *APIError {
	return NewAPIError(409, CodeConflict, message)
}

func RateLimited(message string) *APIError {
	return NewAPIError(429, CodeRateLimited, message)
}

func Invalid(field, message string) error {
	return &ValidationError{Details: []FieldError{{Field: field, Message: message}}}
}

func InvalidFields(details []FieldError) error {
	return &ValidationError{Details: details}
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Code == CodeNotFound
}

func Errorf(status int, code ErrorCode, format string, args ...any) *APIError {
	return NewAPIError(status, code, fmt.Sprintf(format, args...))
}
