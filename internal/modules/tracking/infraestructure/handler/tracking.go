package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	usecases "rea/porticos/internal/modules/tracking/application/use_cases"
	"rea/porticos/internal/modules/tracking/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/middlewares"

	"github.com/gin-gonic/gin"
)

type TrackingHandler struct {
	uc      *usecases.TrackingUseCase
	Version int
}

func NewTrackingHandler(uc *usecases.TrackingUseCase) *TrackingHandler {
	return &TrackingHandler{uc: uc, Version: 1}
}

func (h *TrackingHandler) GetVersion() int {
	return h.Version
}

func (h *TrackingHandler) Position(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var req requests.TrackingPositionRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	result, err := h.uc.ProcessPosition(c.Request.Context(), ownerID, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
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

func decodeStrictJSON(c *gin.Context, target any) error {
	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	if !strings.HasPrefix(contentType, "application/json") {
		return domainErrors.NewValidationError("CONTENT_TYPE_INVALID", "Content-Type debe ser application/json")
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return domainErrors.NewInternalError("REQUEST_BODY_READ_ERROR", "no se pudo leer request body")
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return domainErrors.NewValidationError("REQUEST_BODY_REQUIRED", "body JSON es obligatorio")
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domainErrors.NewValidationError("JSON_INVALID", "JSON inválido o contiene campos no permitidos")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return domainErrors.NewValidationError("JSON_INVALID", "JSON inválido: múltiples objetos no permitidos")
	}
	return nil
}

func respondError(c *gin.Context, err error) {
	status, payload := httpMapper.MapErrorToHttp(err)
	c.JSON(status, payload)
}
