package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	usecases "rea/porticos/internal/modules/accounts/application/use_cases"
	"rea/porticos/internal/modules/accounts/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
	httpMapper "rea/porticos/pkg/http"
	"rea/porticos/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AccountsHandler struct {
	uc      *usecases.AccountsUseCase
	Version int
}

func NewAccountsHandler(uc *usecases.AccountsUseCase) *AccountsHandler {
	return &AccountsHandler{
		uc:      uc,
		Version: 1,
	}
}

func (h *AccountsHandler) GetVersion() int {
	return h.Version
}

func (h *AccountsHandler) CreateManaged(c *gin.Context) {
	var req requests.CreateAccountRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	profile, err := h.uc.CreateAccount(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             profile.ID,
		"supabaseUserId": profile.SupabaseUserID,
		"email":          profile.Email,
		"role":           profile.Role,
		"status":         profile.Status,
	})
}

func (h *AccountsHandler) Signup(c *gin.Context) {
	var req requests.CreateAccountRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	profile, err := h.uc.SignupPublic(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             profile.ID,
		"supabaseUserId": profile.SupabaseUserID,
		"email":          profile.Email,
		"role":           profile.Role,
		"status":         profile.Status,
	})
}

func (h *AccountsHandler) CreateFirstAdmin(c *gin.Context) {
	var req requests.CreateAccountRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	profile, err := h.uc.BootstrapFirstAdmin(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             profile.ID,
		"supabaseUserId": profile.SupabaseUserID,
		"email":          profile.Email,
		"role":           profile.Role,
		"status":         profile.Status,
	})
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
