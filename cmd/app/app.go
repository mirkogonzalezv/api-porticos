package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	configuracion "rea/porticos/cmd/config"
	"rea/porticos/cmd/container"
	"rea/porticos/cmd/routes"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/middlewares"
	"strconv"
	"syscall"
	"time"

	"github.com/danielkov/gin-helmet/ginhelmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	config *configuracion.Configuracion
	router *gin.Engine
	log    *zap.Logger
}

func NewApp() *App {
	return &App{}
}

func (a *App) Initializar() error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	// Inicializamos cargas de variables de entorno
	if err := configuracion.CargarEnv(env); err != nil {
		return err
	}

	// Inicializamos el logger
	logger.Init(env)
	logger.General("=== API Porticos ===")

	cfg, err := configuracion.CargarVariables()
	if err != nil {
		return err
	}

	a.config = cfg
	logger.Success("Configuración cargada correctamente")
	logger.General("Servidor HTTP Configurado")

	router := gin.New()

	// Middleware para activar los logs a los servicios
	router.Use(gin.Logger())

	// Middleware para error handling global
	router.Use(middlewares.ErrorHandlerMiddleware())
	logger.General("Error handler global configurado")

	// Middleare para que panic handler
	router.Use(gin.Recovery())

	// Agregamos middlewares headers de seguridad (OWASP top 10)
	router.Use(ginhelmet.Default())
	logger.Success("Headers de seguridad configurados para entorno: " + env)

	// Agregamos middleware CORS global
	router.Use(cors.Default())
	logger.Success("CORS configurado para entorno: " + env)

	// Después de configurar middlewares
	container := container.NewContainer()
	apiRouter := routes.NewAPIRouter(container)
	apiRouter.RegisterRoutes(router)

	a.router = router

	return nil
}

func (a *App) Run() error {
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(a.config.Port),
		Handler:      a.router,
		ReadTimeout:  time.Duration(a.config.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(a.config.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(a.config.IdleTimeoutSec) * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		logger.Success("Atlas seed iniciado correctamente en puerto " + strconv.Itoa(a.config.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Esperar error de servidor o señal de terminación
	select {
	case err := <-errCh:
		logger.Error("Error al iniciar servidor: " + err.Error())
		return err
	case sig := <-quit:
		logger.General("Señal recibida: " + sig.String() + ", terminando atlas seed...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.ShutDownTimeout)*time.Second)
	defer cancel()

	// Apagado del servidor
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Error durante el apagado..." + err.Error())
		return err
	}

	// Llamar al método para cerrar recursos
	a.Apagar()

	// Asegurar que todos los logs se escriban
	logger.L().Sync()
	logger.Success("Servidor terminado correctamente")
	return nil
}

func (a *App) Apagar() {
	// Cerrar conexiónes a DB u otras dependencias
	logger.General("Cerrando conexiones y recursos...")

	// DB cerrar
	// Cerrar Colas
	// Cerrar recursos

	// Simular tiempo de cierre de recursos
	time.Sleep(100 * time.Millisecond)

	logger.General("Recursos cerrados correctamente")
}
