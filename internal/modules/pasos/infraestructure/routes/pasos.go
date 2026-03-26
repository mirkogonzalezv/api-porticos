package routes

import (
	"rea/porticos/internal/modules/pasos/infraestructure/handler"
	"rea/porticos/pkg/middlewares"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigPasosVersion(pasosHandler *handler.PasosHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		pasosCtrl := ctrl.(*handler.PasosHandler)
		RegisterPasosRoutes(rg, pasosCtrl)
	}

	version.ConfigControllerVersion("pasos", pasosHandler, wrapperFunc)
}

func RegisterPasosRoutes(rg *gin.RouterGroup, h *handler.PasosHandler) {
	allowed := middlewares.RequireRoles("reader", "partner", "admin")
	adminOnly := middlewares.RequireRoles("admin")

	rg.POST("", allowed, h.Create)
	rg.GET("", allowed, h.List)
	rg.GET("/capturados", allowed, h.ListCapturados)
	rg.GET("/admin", adminOnly, h.ListAll)
	rg.GET("/capturados/admin", adminOnly, h.ListCapturadosAll)
	rg.GET("/resumen", allowed, h.Summary)
	rg.GET("/:id", allowed, h.GetByID)
}
