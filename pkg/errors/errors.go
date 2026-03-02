package errors

import "fmt"

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
)

type DomainError struct {
	Type    ErrorType `json:"type"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// WithDetails agrega detalles adicionales al error
func (e *DomainError) WithDetails(details string) *DomainError {
	e.Details = details
	return e
}

// Constructores de errores por tipo
func NewValidationError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    code,
		Message: message,
	}
}

func NewNotFoundError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Code:    code,
		Message: message,
	}
}

func NewConflictError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Code:    code,
		Message: message,
	}
}

func NewUnauthorizedError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Code:    code,
		Message: message,
	}
}

func NewForbiddenError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeForbidden,
		Code:    code,
		Message: message,
	}
}

func NewInternalError(code, message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    code,
		Message: message,
	}
}
