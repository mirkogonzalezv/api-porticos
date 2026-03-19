package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	usecases "rea/porticos/internal/modules/concesionarias/application/use_cases"
	requests "rea/porticos/internal/modules/concesionarias/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ConcesionariasHandler struct {
	uc      *usecases.ConcesionariasUseCase
	Version int
}

func NewConcesionariasHandler(uc *usecases.ConcesionariasUseCase) *ConcesionariasHandler {
	return &ConcesionariasHandler{uc: uc, Version: 1}
}

func (h *ConcesionariasHandler) GetVersion() int {
	return h.Version
}

func (h *ConcesionariasHandler) List(c *gin.Context) {
	limit, err := parseQueryInt(c, "limit", 20)
	if err != nil {
		respondError(c, err)
		return
	}
	offset, err := parseQueryInt(c, "offset", 0)
	if err != nil {
		respondError(c, err)
		return
	}

	estado := strings.TrimSpace(c.Query("estado"))

	items, err := h.uc.List(c.Request.Context(), limit, offset, estado)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *ConcesionariasHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	out, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ConcesionariasHandler) Create(c *gin.Context) {
	var req requests.ConcesionariaUpsertRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	entity, err := req.ToEntity()
	if err != nil {
		respondError(c, err)
		return
	}
	out, err := h.uc.Create(c.Request.Context(), entity)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *ConcesionariasHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req requests.ConcesionariaUpsertRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	entity, err := req.ToEntityWithID(id)
	if err != nil {
		respondError(c, err)
		return
	}
	out, err := h.uc.Update(c.Request.Context(), entity)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *ConcesionariasHandler) Delete(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseQueryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, domainErrors.NewValidationError("QUERY_INVALID", key+" debe ser un entero válido")
	}
	if value < 0 {
		return 0, domainErrors.NewValidationError("QUERY_INVALID", key+" no puede ser negativo")
	}
	return value, nil
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
	_ = c.Error(err)
	status, payload := httpMapper.MapErrorToHttp(err)
	if status >= 500 {
		logger.L().Error("Handler error",
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Int("status", status),
			zap.Error(err),
		)
	}
	c.JSON(status, payload)
}
