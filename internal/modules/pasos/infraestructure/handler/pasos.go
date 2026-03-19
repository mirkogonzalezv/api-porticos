package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	usecases "rea/porticos/internal/modules/pasos/application/use_cases"
	requests "rea/porticos/internal/modules/pasos/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PasosHandler struct {
	uc      *usecases.PasosUseCase
	Version int
}

func NewPasosHandler(uc *usecases.PasosUseCase) *PasosHandler {
	return &PasosHandler{uc: uc, Version: 1}
}

func (h *PasosHandler) GetVersion() int {
	return h.Version
}

func (h *PasosHandler) Create(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var req requests.CreatePasoPorticoRequest
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

func (h *PasosHandler) CreateBatch(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	var req requests.CreatePasoBatchRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	if err := req.Validate(); err != nil {
		respondError(c, err)
		return
	}

	items, err := h.uc.CreateBatch(c.Request.Context(), ownerID, req.Items)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":  items,
		"count": len(items),
	})
}

func (h *PasosHandler) GetByID(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	out, err := h.uc.GetByID(c.Request.Context(), ownerID, id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

func (h *PasosHandler) List(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	from, to, err := parseRangeQuery(c)
	if err != nil {
		respondError(c, err)
		return
	}

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

	vehiculoID := strings.TrimSpace(c.Query("vehiculoId"))
	porticoID := strings.TrimSpace(c.Query("porticoId"))

	items, err := h.uc.ListByOwnerRange(c.Request.Context(), ownerID, from, to, vehiculoID, porticoID, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"limit":      limit,
		"offset":     offset,
		"from":       from.Format(time.RFC3339),
		"to":         to.Format(time.RFC3339),
		"vehiculoId": vehiculoID,
		"porticoId":  porticoID,
	})
}

func (h *PasosHandler) ListAll(c *gin.Context) {
	from, to, err := parseRangeQuery(c)
	if err != nil {
		respondError(c, err)
		return
	}

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

	vehiculoID := strings.TrimSpace(c.Query("vehiculoId"))
	porticoID := strings.TrimSpace(c.Query("porticoId"))

	items, err := h.uc.ListAllRange(c.Request.Context(), from, to, vehiculoID, porticoID, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"limit":      limit,
		"offset":     offset,
		"from":       from.Format(time.RFC3339),
		"to":         to.Format(time.RFC3339),
		"vehiculoId": vehiculoID,
		"porticoId":  porticoID,
	})
}

func (h *PasosHandler) Summary(c *gin.Context) {
	ownerID, err := getAuthUserID(c)
	if err != nil {
		respondError(c, err)
		return
	}

	from, to, err := parseRangeQuery(c)
	if err != nil {
		respondError(c, err)
		return
	}

	vehiculoID := strings.TrimSpace(c.Query("vehiculoId"))
	porticoID := strings.TrimSpace(c.Query("porticoId"))
	groupBy := strings.TrimSpace(c.Query("groupBy"))
	if groupBy == "" {
		groupBy = "day"
	}

	items, err := h.uc.SummaryByOwnerRange(c.Request.Context(), ownerID, from, to, vehiculoID, porticoID, groupBy)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       items,
		"groupBy":    strings.ToLower(groupBy),
		"from":       from.Format(time.RFC3339),
		"to":         to.Format(time.RFC3339),
		"vehiculoId": vehiculoID,
		"porticoId":  porticoID,
	})
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

func parseRangeQuery(c *gin.Context) (time.Time, time.Time, error) {
	fromRaw := strings.TrimSpace(c.Query("from"))
	toRaw := strings.TrimSpace(c.Query("to"))
	if fromRaw == "" || toRaw == "" {
		return time.Time{}, time.Time{}, domainErrors.NewValidationError("PASO_RANGE_REQUIRED", "from y to son obligatorios")
	}

	from, err := parseDateTime(fromRaw, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := parseDateTime(toRaw, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, domainErrors.NewValidationError("PASO_RANGE_INVALID", "to no puede ser menor que from")
	}

	return from, to, nil
}

func parseDateTime(input string, endOfDay bool) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, domainErrors.NewValidationError("PASO_FECHA_INVALID", "fecha inválida")
	}

	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	if d, err := time.Parse("2006-01-02", input); err == nil {
		if endOfDay {
			return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC), nil
		}
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, domainErrors.NewValidationError("PASO_FECHA_INVALID", "usa RFC3339 o YYYY-MM-DD")
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
