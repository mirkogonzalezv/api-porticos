package container

import (
	healthUseCase "rea/porticos/internal/modules/health/application/use_case"
	healthController "rea/porticos/internal/modules/health/infrastructure/controller"
	healthRoutes "rea/porticos/internal/modules/health/infrastructure/routes"
	"rea/porticos/pkg/logger"	
)

// Container para manejo de inyección de dependencias
type Container struct {
	HealthController *healthController.HealthController
}

func NewContainer() *Container {
	healthUseCase := healthUseCase.NewHealthUseCase()
	healthController := healthController.NewHealthController(healthUseCase)
	// Se configura el versionado automático
	healthRoutes.ConfigHealthVersion(healthController)

	logger.Success("Container de dependencias inicializado...")

	return &Container{
		HealthController: healthController,
	}
}
