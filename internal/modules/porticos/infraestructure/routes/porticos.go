package routes

import (
	"rea/porticos/internal/modules/porticos/infraestructure/handler"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigPorticosVersion(porticosHandler *handler.PorticosHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		porticosCtrl := ctrl.(*handler.PorticosHandler)
		RegisterPorticosRoutes(rg, porticosCtrl)
	}

	version.ConfigControllerVersion("porticos", porticosHandler, wrapperFunc)
}

func RegisterPorticosRoutes(rg *gin.RouterGroup, h *handler.PorticosHandler) {
	rg.GET("", h.List)
	rg.GET("/:id", h.GetByID)
	rg.POST("", h.Create)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}
