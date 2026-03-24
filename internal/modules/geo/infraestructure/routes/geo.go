package routes

import (
	"rea/porticos/internal/modules/geo/infraestructure/handler"
	"rea/porticos/pkg/middlewares"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigGeoVersion(geoHandler *handler.GeoBatchHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		geoCtrl := ctrl.(*handler.GeoBatchHandler)
		RegisterGeoRoutes(rg, geoCtrl)
	}

	version.ConfigControllerVersion("geo", geoHandler, wrapperFunc)
}

func RegisterGeoRoutes(rg *gin.RouterGroup, h *handler.GeoBatchHandler) {
	authRoles := middlewares.RequireRoles("reader", "partner", "admin")

	rg.POST("/batch", authRoles, h.IngestBatch)
}
