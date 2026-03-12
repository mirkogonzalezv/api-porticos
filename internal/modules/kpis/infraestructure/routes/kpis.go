package routes

import (
	"rea/porticos/internal/modules/kpis/infraestructure/handler"
	"rea/porticos/pkg/middlewares"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

func ConfigKPIsVersion(kpisHandler *handler.KPIsHandler) {
	wrapperFunc := func(rg *gin.RouterGroup, ctrl any) {
		kpisCtrl := ctrl.(*handler.KPIsHandler)
		RegisterKPIsRoutes(rg, kpisCtrl)
	}

	version.ConfigControllerVersion("kpis", kpisHandler, wrapperFunc)
}

func RegisterKPIsRoutes(rg *gin.RouterGroup, h *handler.KPIsHandler) {
	adminOnly := middlewares.RequireRoles("admin")
	rg.GET("", adminOnly, h.GetBasic)
}
