package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	configuracion "rea/porticos/cmd/config"
	"rea/porticos/cmd/container"
	"rea/porticos/cmd/routes"
	"rea/porticos/pkg/db"
	"rea/porticos/pkg/logger"
	"rea/porticos/pkg/middlewares"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	config *configuracion.Configuracion
	router *gin.Engine
	log    *zap.Logger
	db     *db.Postgres
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

	if strings.TrimSpace(cfg.SupabaseJWKSURL) == "" ||
		strings.TrimSpace(cfg.SupabaseJWTIssuer) == "" ||
		strings.TrimSpace(cfg.SupabaseJWTAudience) == "" {
		return fmt.Errorf("faltan variables SUPABASE_JWKS_URL, SUPABASE_JWT_ISSUER o SUPABASE_JWT_AUDIENCE")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pg, err := db.NewPostgres(
		ctx,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	if err != nil {
		return fmt.Errorf("db init failed: %w", err)
	}

	a.db = pg
	a.config = cfg
	logger.Success("Conexión a PostgreSQL validada")
	logger.Success("Configuración cargada correctamente")
	logger.General("Servidor HTTP Configurado")

	router := gin.New()

	middlewares.Register(router, middlewares.Options{
		Environment:         env,
		AllowedOrigins:      cfg.AllowedOrigins,
		SupabaseJWKSURL:     cfg.SupabaseJWKSURL,
		SupabaseJWTIssuer:   cfg.SupabaseJWTIssuer,
		SupabaseJWTAudience: cfg.SupabaseJWTAudience,
	})

	// Después de configurar middlewares
	container := container.NewContainer(a.db)
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

	if a.db != nil {
		a.db.Close()
		logger.General("Conexión PostgreSQL cerrada")
	}

	// Simular tiempo de cierre de recursos
	time.Sleep(100 * time.Millisecond)

	logger.General("Recursos cerrados correctamente")
}
