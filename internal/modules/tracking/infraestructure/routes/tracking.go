package routes

import (
	"rea/porticos/internal/modules/tracking/infraestructure/handler"
	"rea/porticos/pkg/middlewares"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigTrackingVersion(trackingHandler *handler.TrackingHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		trackingCtrl := ctrl.(*handler.TrackingHandler)
		RegisterTrackingRoutes(rg, trackingCtrl)
	}

	version.ConfigControllerVersion("tracking", trackingHandler, wrapperFunc)
}

func RegisterTrackingRoutes(rg *gin.RouterGroup, h *handler.TrackingHandler) {
	authRoles := middlewares.RequireRoles("reader", "partner", "admin")

	rg.POST("/position", authRoles, h.Position)
}
