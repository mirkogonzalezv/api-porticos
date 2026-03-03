package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
)

type Configuracion struct {
	AppEnv          string `env:"ENVIRONMENT" envDefault:"dev" validate:"oneof=dev qa production"`
	Port            int    `env:"PORT" envDefault:"3200"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	ShutDownTimeout int    `env:"SHUTDOWN_TIMEOUT" envDefault:"30"`
	// Se definen las variables de entorno necesarias para el proyecto
	ReadTimeoutSec  int `env:"READ_TIMEOUT_SEC" envDefault:"10" validate:"gte=1,lte=300"`
	WriteTimeoutSec int `env:"WRITE_TIMEOUT_SEC" envDefault:"15" validate:"gte=1,lte=300"`
	IdleTimeoutSec  int `env:"IDLE_TIMEOUT_SEC" envDefault:"60" validate:"gte=1,lte=600"`

	BodyMaxBytes       int64  `env:"BODY_MAX_BYTES" envDefault:"1048576" validate:"gte=1024,lte=10485760"`
	RateLimit          int    `env:"RATE_LIMIT" envDefault:"100" validate:"gte=1,lte=10000"`
	RateLimitWindowSec int    `env:"RATE_LIMIT_WINDOW_SEC" envDefault:"60" validate:"gte=1,lte=3600"`
	AllowedOrigins     string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000"`
}

func CargarVariables() (*Configuracion, error) {
	cfg := &Configuracion{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	v := validator.New()

	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
