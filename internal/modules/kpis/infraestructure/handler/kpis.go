package handler

import (
	"net/http"

	usecases "rea/porticos/internal/modules/kpis/application/use_cases"
	httpMapper "rea/porticos/pkg/http"

	"github.com/gin-gonic/gin"
)

type KPIsHandler struct {
	uc      *usecases.KPIsUseCase
	Version int
}

func NewKPIsHandler(uc *usecases.KPIsUseCase) *KPIsHandler {
	return &KPIsHandler{uc: uc, Version: 1}
}

func (h *KPIsHandler) GetVersion() int {
	return h.Version
}

func (h *KPIsHandler) GetBasic(c *gin.Context) {
	out, err := h.uc.GetBasicKPIs(c.Request.Context())
	if err != nil {
		status, payload := httpMapper.MapErrorToHttp(err)
		c.JSON(status, payload)
		return
	}

	c.JSON(http.StatusOK, out)
}
