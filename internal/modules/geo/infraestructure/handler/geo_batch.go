package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	usecases "rea/porticos/internal/modules/geo/application/use_cases"
	"rea/porticos/internal/modules/geo/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GeoBatchHandler struct {
	uc      *usecases.GeoBatchUseCase
	Version int
}

func NewGeoBatchHandler(uc *usecases.GeoBatchUseCase) *GeoBatchHandler {
	return &GeoBatchHandler{uc: uc, Version: 1}
}

func (h *GeoBatchHandler) GetVersion() int {
	return h.Version
}

func (h *GeoBatchHandler) IngestBatch(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var req requests.GeoBatchRequest
	rawBody, err := decodeStrictJSON(c, &req)
	if err != nil {
		logInvalidPayload(c, rawBody, err)
		respondError(c, err)
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
	}
	result, err := h.uc.ProcessBatch(c.Request.Context(), ownerID, &req, idempotencyKey)
	if err != nil {
		respondError(c, err)
		return
	}

	_ = result
	c.Status(http.StatusCreated)
}

func getAuthUserID(c *gin.Context) (string, error) {
	raw, ok := c.Get(middlewares.ContextUserIDKey)
	if !ok {
		return "", domainErrors.NewUnauthorizedError("AUTH_REQUIRED", "usuario no autenticado")
	}
	userID, ok := raw.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", domainErrors.NewUnauthorizedError("AUTH_REQUIRED", "usuario no autenticado")
	}
	return strings.TrimSpace(userID), nil
}

func decodeStrictJSON(c *gin.Context, target any) (string, error) {
	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	if !strings.HasPrefix(contentType, "application/json") {
		return "", domainErrors.NewValidationError("CONTENT_TYPE_INVALID", "Content-Type debe ser application/json")
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", domainErrors.NewInternalError("REQUEST_BODY_READ_ERROR", "no se pudo leer request body")
	}
	raw := strings.TrimSpace(string(body))
	if len(raw) == 0 {
		return "", domainErrors.NewValidationError("REQUEST_BODY_REQUIRED", "body JSON es obligatorio")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return raw, domainErrors.NewValidationError("JSON_INVALID", "JSON inválido: "+err.Error())
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return raw, domainErrors.NewValidationError("JSON_INVALID", "JSON inválido: múltiples objetos no permitidos")
	}
	return raw, nil
}

func respondError(c *gin.Context, err error) {
	status, payload := httpMapper.MapErrorToHttp(err)
	c.JSON(status, payload)
}

func logInvalidPayload(c *gin.Context, raw string, err error) {
	const maxLen = 2048
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) > maxLen {
		trimmed = trimmed[:maxLen] + "...(truncated)"
	}
	logger.Error("Geo batch payload inválido",
		zap.Error(err),
		zap.String("user_id", c.GetString(middlewares.ContextUserIDKey)),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("payload", trimmed),
	)
}
