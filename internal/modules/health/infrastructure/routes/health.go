package routes

import (
	"rea/porticos/internal/modules/health/infrastructure/controller"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

// Estructura para generar el base path del modulo
func ConfigHealthVersion(healthController *controller.HealthController) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		healthCtrl := ctrl.(*controller.HealthController)
		RegisterHealthRoutes(rg, healthCtrl)
	}

	version.ConfigControllerVersion("health", healthController, wrapperFunc)
}

func RegisterHealthRoutes(rg *gin.RouterGroup, healthController *controller.HealthController) {
	rg.GET("", healthController.GetHealth)

}
