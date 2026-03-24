package container

import (
	"fmt"
	configuracion "rea/porticos/cmd/config"
	accountsData "rea/porticos/internal/modules/accounts/application/data"
	accountsUseCases "rea/porticos/internal/modules/accounts/application/use_cases"
	accountsHandler "rea/porticos/internal/modules/accounts/infraestructure/handler"
	accountsRoutes "rea/porticos/internal/modules/accounts/infraestructure/routes"
	concesionariasData "rea/porticos/internal/modules/concesionarias/application/data"
	concesionariasUseCases "rea/porticos/internal/modules/concesionarias/application/use_cases"
	concesionariasHandler "rea/porticos/internal/modules/concesionarias/infraestructure/handler"
	concesionariasRoutes "rea/porticos/internal/modules/concesionarias/infraestructure/routes"
	geoUseCases "rea/porticos/internal/modules/geo/application/use_cases"
	geoHandler "rea/porticos/internal/modules/geo/infraestructure/handler"
	geoRoutes "rea/porticos/internal/modules/geo/infraestructure/routes"
	healthUseCase "rea/porticos/internal/modules/health/application/use_case"
	healthController "rea/porticos/internal/modules/health/infrastructure/controller"
	healthRoutes "rea/porticos/internal/modules/health/infrastructure/routes"
	kpisData "rea/porticos/internal/modules/kpis/application/data"
	kpisUseCases "rea/porticos/internal/modules/kpis/application/use_cases"
	kpisHandler "rea/porticos/internal/modules/kpis/infraestructure/handler"
	kpisRoutes "rea/porticos/internal/modules/kpis/infraestructure/routes"
	pasosData "rea/porticos/internal/modules/pasos/application/data"
	pasosUseCases "rea/porticos/internal/modules/pasos/application/use_cases"
	pasosHandler "rea/porticos/internal/modules/pasos/infraestructure/handler"
	pasosRoutes "rea/porticos/internal/modules/pasos/infraestructure/routes"
	porticosData "rea/porticos/internal/modules/porticos/application/data"
	porticosUseCases "rea/porticos/internal/modules/porticos/application/use_cases"
	porticosHandler "rea/porticos/internal/modules/porticos/infraestructure/handler"
	porticosRoutes "rea/porticos/internal/modules/porticos/infraestructure/routes"
	trackingData "rea/porticos/internal/modules/tracking/application/data"
	trackingUseCases "rea/porticos/internal/modules/tracking/application/use_cases"
	trackingHandler "rea/porticos/internal/modules/tracking/infraestructure/handler"
	trackingRoutes "rea/porticos/internal/modules/tracking/infraestructure/routes"
	vehiculosData "rea/porticos/internal/modules/vehiculos/application/data"
	vehiculosUseCases "rea/porticos/internal/modules/vehiculos/application/use_cases"
	vehiculosHandler "rea/porticos/internal/modules/vehiculos/infraestructure/handler"
	vehiculosRoutes "rea/porticos/internal/modules/vehiculos/infraestructure/routes"
	"rea/porticos/pkg/cache"
	"rea/porticos/pkg/db"
	"rea/porticos/pkg/logger"
	"strings"
	"time"
)

// Container para manejo de inyección de dependencias
type Container struct {
	HealthController         *healthController.HealthController
	PorticosController       *porticosHandler.PorticosHandler
	AccountsController       *accountsHandler.AccountsHandler
	VehiculosController      *vehiculosHandler.VehiculosHandler
	PasosController          *pasosHandler.PasosHandler
	KPIsController           *kpisHandler.KPIsHandler
	ConcesionariasController *concesionariasHandler.ConcesionariasHandler
	TrackingController       *trackingHandler.TrackingHandler
	GeoController            *geoHandler.GeoBatchHandler
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

	kpisRepo := kpisData.NewKPIsPostgresRepository(dbConn.Pool)
	kpisUseCase := kpisUseCases.NewKPIsUseCase(kpisRepo)
	kpisController := kpisHandler.NewKPIsHandler(kpisUseCase)
	kpisRoutes.ConfigKPIsVersion(kpisController)

	concesionariasRepo := concesionariasData.NewConcesionariasPostgresRepository(dbConn.Pool)
	concesionariasUseCase := concesionariasUseCases.NewConcesionariasUseCase(concesionariasRepo)
	concesionariasController := concesionariasHandler.NewConcesionariasHandler(concesionariasUseCase)
	concesionariasRoutes.ConfigConcesionariasVersion(concesionariasController)

	geoUseCase := geoUseCases.NewGeoBatchUseCase()
	geoController := geoHandler.NewGeoBatchHandler(geoUseCase)
	geoRoutes.ConfigGeoVersion(geoController)

	var trackingStore trackingData.TrackingStore = trackingData.NewTrackingMemoryRepository()
	if strings.TrimSpace(cfg.RedisHost) != "" {
		redisClient, err := cache.NewRedis(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDB, cfg.RedisSSL)
		if err != nil {
			logger.Error("Error conectando Redis, usando memoria: " + err.Error())
		} else {
			trackingStore = trackingData.NewTrackingRedisRepository(redisClient.Client)
			logger.Success("Redis conectado correctamente")
			logger.General("Redis host=" + cfg.RedisHost + " port=" + fmt.Sprintf("%d", cfg.RedisPort) + " db=" + fmt.Sprintf("%d", cfg.RedisDB) + " ssl=" + fmt.Sprintf("%t", cfg.RedisSSL))
		}
	} else {
		logger.General("REDIS_HOST vacío, tracking en memoria")
	}
	trackingUseCase := trackingUseCases.NewTrackingUseCase(porticosRepo, vehiculosRepo, pasosRepo, trackingStore)
	trackingController := trackingHandler.NewTrackingHandler(trackingUseCase)
	trackingRoutes.ConfigTrackingVersion(trackingController)

	logger.Success("Container de dependencias inicializado...")

	return &Container{
		HealthController:         healthController,
		PorticosController:       porticosController,
		AccountsController:       accountsController,
		VehiculosController:      vehiculosController,
		PasosController:          pasosController,
		KPIsController:           kpisController,
		ConcesionariasController: concesionariasController,
		TrackingController:       trackingController,
		GeoController:            geoController,
	}
}
