package container

import (
	healthUseCase "rea/porticos/internal/modules/health/application/use_case"
	healthController "rea/porticos/internal/modules/health/infrastructure/controller"
	healthRoutes "rea/porticos/internal/modules/health/infrastructure/routes"
	porticosData "rea/porticos/internal/modules/porticos/application/data"
	porticosUseCases "rea/porticos/internal/modules/porticos/application/use_cases"
	porticosHandler "rea/porticos/internal/modules/porticos/infraestructure/handler"
	porticosRoutes "rea/porticos/internal/modules/porticos/infraestructure/routes"
	"rea/porticos/pkg/db"
	"rea/porticos/pkg/logger"
)

// Container para manejo de inyección de dependencias
type Container struct {
	HealthController   *healthController.HealthController
	PorticosController *porticosHandler.PorticosHandler
}

func NewContainer(dbConn *db.Postgres) *Container {
	healthUseCase := healthUseCase.NewHealthUseCase()
	healthController := healthController.NewHealthController(healthUseCase)
	// Se configura el versionado automático
	healthRoutes.ConfigHealthVersion(healthController)

	porticosRepo := porticosData.NewPostgresPorticoRepository(dbConn.Pool)
	porticosUseCase := porticosUseCases.NewPorticosUseCase(porticosRepo)
	porticosController := porticosHandler.NewPorticosHandler(porticosUseCase)
	porticosRoutes.ConfigPorticosVersion(porticosController)

	logger.Success("Container de dependencias inicializado...")

	return &Container{
		HealthController:   healthController,
		PorticosController: porticosController,
	}
}
