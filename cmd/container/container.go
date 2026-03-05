package container

import (
	configuracion "rea/porticos/cmd/config"
	accountsData "rea/porticos/internal/modules/accounts/application/data"
	accountsUseCases "rea/porticos/internal/modules/accounts/application/use_cases"
	accountsHandler "rea/porticos/internal/modules/accounts/infraestructure/handler"
	accountsRoutes "rea/porticos/internal/modules/accounts/infraestructure/routes"
	healthUseCase "rea/porticos/internal/modules/health/application/use_case"
	healthController "rea/porticos/internal/modules/health/infrastructure/controller"
	healthRoutes "rea/porticos/internal/modules/health/infrastructure/routes"
	pasosData "rea/porticos/internal/modules/pasos/application/data"
	pasosUseCases "rea/porticos/internal/modules/pasos/application/use_cases"
	pasosHandler "rea/porticos/internal/modules/pasos/infraestructure/handler"
	pasosRoutes "rea/porticos/internal/modules/pasos/infraestructure/routes"
	porticosData "rea/porticos/internal/modules/porticos/application/data"
	porticosUseCases "rea/porticos/internal/modules/porticos/application/use_cases"
	porticosHandler "rea/porticos/internal/modules/porticos/infraestructure/handler"
	porticosRoutes "rea/porticos/internal/modules/porticos/infraestructure/routes"
	vehiculosData "rea/porticos/internal/modules/vehiculos/application/data"
	vehiculosUseCases "rea/porticos/internal/modules/vehiculos/application/use_cases"
	vehiculosHandler "rea/porticos/internal/modules/vehiculos/infraestructure/handler"
	vehiculosRoutes "rea/porticos/internal/modules/vehiculos/infraestructure/routes"
	"rea/porticos/pkg/db"
	"rea/porticos/pkg/logger"
	"time"
)

// Container para manejo de inyección de dependencias
type Container struct {
	HealthController    *healthController.HealthController
	PorticosController  *porticosHandler.PorticosHandler
	AccountsController  *accountsHandler.AccountsHandler
	VehiculosController *vehiculosHandler.VehiculosHandler
	PasosController     *pasosHandler.PasosHandler
}

func NewContainer(dbConn *db.Postgres, cfg *configuracion.Configuracion) *Container {
	healthUseCase := healthUseCase.NewHealthUseCase()
	healthController := healthController.NewHealthController(healthUseCase)
	// Se configura el versionado automático
	healthRoutes.ConfigHealthVersion(healthController)

	porticosRepo := porticosData.NewPostgresPorticoRepository(dbConn.Pool)
	porticosRepo = porticosData.NewCachedPorticoRepository(porticosRepo, 20*time.Second, 300)
	porticosUseCase := porticosUseCases.NewPorticosUseCase(porticosRepo)
	porticosController := porticosHandler.NewPorticosHandler(porticosUseCase)
	porticosRoutes.ConfigPorticosVersion(porticosController)

	profilesRepo := accountsData.NewProfilePostgresRepository(dbConn.Pool)
	supabaseClient := accountsData.NewSupabaseAdminClient(cfg.SupabaseURL, cfg.SupabaseServiceRole)
	accountsUseCase := accountsUseCases.NewAccountsUseCase(profilesRepo, supabaseClient)
	accountsController := accountsHandler.NewAccountsHandler(accountsUseCase)
	accountsRoutes.ConfigAccountsVersion(accountsController)

	vehiculosRepo := vehiculosData.NewVehiculosPostgresRepository(dbConn.Pool)
	vehiculosUseCase := vehiculosUseCases.NewVehiculosUseCase(vehiculosRepo)
	vehiculosController := vehiculosHandler.NewVehiculosHandler(vehiculosUseCase)
	vehiculosRoutes.ConfigVehiculosVersion(vehiculosController)

	pasosRepo := pasosData.NewPasosPostgresRepository(dbConn.Pool)
	pasosUseCase := pasosUseCases.NewPasosUseCase(pasosRepo, vehiculosRepo, porticosRepo)
	pasosController := pasosHandler.NewPasosHandler(pasosUseCase)
	pasosRoutes.ConfigPasosVersion(pasosController)

	logger.Success("Container de dependencias inicializado...")

	return &Container{
		HealthController:    healthController,
		PorticosController:  porticosController,
		AccountsController:  accountsController,
		VehiculosController: vehiculosController,
		PasosController:     pasosController,
	}
}
