package routes

import (
	"rea/porticos/cmd/container"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/version"

	"github.com/gin-gonic/gin"
)

type APIRouter struct {
	container *container.Container
}

func NewAPIRouter(cont *container.Container) *APIRouter {
	return &APIRouter{
		container: cont,
	}
}

// RegisterRoutes registra todas las rutas de la aplicación
func (r *APIRouter) RegisterRoutes(router *gin.Engine) {

	logger.Success("RUTA API registradas correctamente")
	// Segundo parametro recibe path anterior a versión
	version.BuildRoutes(router, "api")

}
