package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	usecases "rea/porticos/internal/modules/vehiculos/application/use_cases"
	requests "rea/porticos/internal/modules/vehiculos/domain/dtos/requests"
	"rea/porticos/internal/modules/vehiculos/domain/entities"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type VehiculosHandler struct {
	uc      *usecases.VehiculosUseCase
	Version int
}

func NewVehiculosHandler(uc *usecases.VehiculosUseCase) *VehiculosHandler {
	return &VehiculosHandler{uc: uc, Version: 1}
}

func (h *VehiculosHandler) GetVersion() int {
	return h.Version
}

func (h *VehiculosHandler) List(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}
	role := getAuthRole(c)
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

	var items []entities.Vehiculo
	if role == "admin" {
		items, err = h.uc.ListAll(c.Request.Context(), limit, offset)
	} else {
		items, err = h.uc.ListByOwner(c.Request.Context(), ownerID, limit, offset)
	}
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

func (h *VehiculosHandler) GetByID(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}
	role := getAuthRole(c)
	id := strings.TrimSpace(c.Param("id"))
	var out *entities.Vehiculo
	if role == "admin" {
		out, err = h.uc.GetByIDAny(c.Request.Context(), id)
	} else {
		out, err = h.uc.GetByID(c.Request.Context(), ownerID, id)
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *VehiculosHandler) Create(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}
	var req requests.CreateVehiculoRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	entity, err := req.ToEntity(ownerID)
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

func (h *VehiculosHandler) Update(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	var req requests.UpdateVehiculoRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	entity, err := req.ToEntity(ownerID, id)
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

func (h *VehiculosHandler) Delete(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if err := h.uc.Delete(c.Request.Context(), ownerID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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

func getAuthRole(c *gin.Context) string {
	raw, ok := c.Get(middlewares.ContextUserRoleKey)
	if !ok {
		return ""
	}
	role, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(role))
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
