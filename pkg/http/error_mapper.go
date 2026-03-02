package http

import (
	"net/http"
	domainErrors "rea/porticos/pkg/errors"
	"time"
)

type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	Timestamp string      `json:"timestamp"`
}

type ErrorDetail struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MapErrorToHttp convierte errores de dominio a respuestas HTTP
func MapErrorToHttp(err error) (int, ErrorResponse) {
	if domainErr, ok := err.(*domainErrors.DomainError); ok {
		statusCode := getHttpStatusCode(domainErr.Type)

		return statusCode, ErrorResponse{
			Error: ErrorDetail{
				Type:    string(domainErr.Type),
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Details: domainErr.Details,
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Error genérico no controlado
	return http.StatusInternalServerError, ErrorResponse{
		Error: ErrorDetail{
			Type:    string(domainErrors.ErrorTypeInternal),
			Code:    "INTERNAL_001",
			Message: "Error interno del servidor",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func getHttpStatusCode(errorType domainErrors.ErrorType) int {
	switch errorType {
	case domainErrors.ErrorTypeValidation:
		return http.StatusBadRequest
	case domainErrors.ErrorTypeNotFound:
		return http.StatusNotFound
	case domainErrors.ErrorTypeConflict:
		return http.StatusConflict
	case domainErrors.ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case domainErrors.ErrorTypeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
