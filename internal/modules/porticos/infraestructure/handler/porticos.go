package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	usecases "rea/porticos/internal/modules/porticos/application/use_cases"
	requests "rea/porticos/internal/modules/porticos/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"

	"github.com/gin-gonic/gin"
)

type PorticosHandler struct {
	uc      *usecases.PorticosUseCase
	Version int
}

func NewPorticosHandler(uc *usecases.PorticosUseCase) *PorticosHandler {
	return &PorticosHandler{
		uc:      uc,
		Version: 1,
	}
}

func (h *PorticosHandler) GetVersion() int {
	return h.Version
}

func (h *PorticosHandler) List(c *gin.Context) {
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

	result, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	payload := gin.H{
		"data":   result,
		"limit":  limit,
		"offset": offset,
	}
	writeCachedJSON(c, http.StatusOK, payload, 30)
}

func (h *PorticosHandler) GetByID(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	result, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	writeCachedJSON(c, http.StatusOK, result, 30)
}

func (h *PorticosHandler) Create(c *gin.Context) {
	var req requests.PorticoUpsertRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	entity, err := req.ToEntity()
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.uc.Create(c.Request.Context(), entity)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *PorticosHandler) Update(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	var req requests.PorticoUpsertRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	entity, err := req.ToEntity()
	if err != nil {
		respondError(c, err)
		return
	}

	result, err := h.uc.Update(c.Request.Context(), id, entity)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PorticosHandler) Delete(c *gin.Context) {
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
	status, payload := httpMapper.MapErrorToHttp(err)
	c.JSON(status, payload)
}

func writeCachedJSON(c *gin.Context, status int, payload any, maxAgeSec int) {
	raw, err := json.Marshal(payload)
	if err != nil {
		c.JSON(status, payload)
		return
	}

	sum := sha256.Sum256(raw)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	ifNoneMatch := strings.TrimSpace(c.GetHeader("If-None-Match"))
	if ifNoneMatch != "" && ifNoneMatch == etag {
		c.Header("ETag", etag)
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSec)+", stale-while-revalidate=60")
		c.Status(http.StatusNotModified)
		return
	}

	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSec)+", stale-while-revalidate=60")
	c.Data(status, "application/json; charset=utf-8", raw)
}
