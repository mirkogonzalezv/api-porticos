package main

import (
	"rea/porticos/cmd/app"
	"log"
)

// @title Porticos API
// @version 1.0
// @description Microservice about porticos in Chile
// @contact.name Mirko Gonzalez
// @contact.email gonzalez.mirko91@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:4200
// @BasePath /api
func main() {
	applicacion := app.NewApp()

	if err := applicacion.Initializar(); err != nil {
		log.Fatal("Error inicializando aplicación:", err)
	}

	if err := applicacion.Run(); err != nil {
		log.Fatal("Error ejecutando aplicación: ", err)
	}
}
