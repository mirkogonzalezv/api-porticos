package routes

import (
	"rea/porticos/internal/modules/concesionarias/infraestructure/handler"
	"rea/porticos/pkg/middlewares"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigConcesionariasVersion(concesionariasHandler *handler.ConcesionariasHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		concesionariasCtrl := ctrl.(*handler.ConcesionariasHandler)
		RegisterConcesionariasRoutes(rg, concesionariasCtrl)
	}

	version.ConfigControllerVersion("concesionarias", concesionariasHandler, wrapperFunc)
}

func RegisterConcesionariasRoutes(rg *gin.RouterGroup, h *handler.ConcesionariasHandler) {
	adminOnly := middlewares.RequireRoles("admin")

	rg.GET("", adminOnly, h.List)
	rg.GET("/:id", adminOnly, h.GetByID)
	rg.POST("", adminOnly, h.Create)
	rg.PUT("/:id", adminOnly, h.Update)
	rg.DELETE("/:id", adminOnly, h.Delete)
}
