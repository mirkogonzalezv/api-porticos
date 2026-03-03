package controller

import (
	usecase "rea/porticos/internal/modules/health/application/use_case"
	"rea/porticos/internal/modules/health/domain/dto/responses"

	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	uc      *usecase.HealthUseCase
	Version int
}

func NewHealthController(uc *usecase.HealthUseCase) *HealthController {
	return &HealthController{
		uc:      uc,
		Version: 1,
	}
}

// @Summary Health Check
// @Description Returns the health status of the service
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} responses.HealthResponse
// @Failure 503 {object} responses.ErrorResponse
// @Router /health [get]
func (ctr *HealthController) GetHealth(c *gin.Context) {
	response, err := ctr.uc.CheckHealth()
	if err != nil {
		errorResponse := responses.NewErrorResponse("Service unhealthy")
		c.JSON(http.StatusServiceUnavailable, errorResponse)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ctr *HealthController) GetVersion() int {
	return ctr.Version
}
