package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func CargarEnv(environment string) error {
	appEnv := environment
	if appEnv == "" {
		appEnv = os.Getenv("ENVIRONMENT")
	}
	if appEnv == "" {
		appEnv = "dev"
	}

	// Deja consistente el valor para el resto de la app.
	if err := os.Setenv("ENVIRONMENT", appEnv); err != nil {
		return fmt.Errorf("no se pudo setear ENVIRONMENT: %w", err)
	}

	if appEnv == "dev" || appEnv == "qa" {
		filename := ".env." + appEnv

		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("Cargando %s...\n", filename)
			if err := godotenv.Load(filename); err != nil {
				return fmt.Errorf("error cargando %s: %w", filename, err)
			}
		} else {
			fmt.Printf("Archivo %s no encontrado, usando variables del sistema\n", filename)
		}
	}

	fmt.Println("ENVIRONMENT:", appEnv)
	return nil
}
