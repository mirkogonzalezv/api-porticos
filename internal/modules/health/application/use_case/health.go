package usecase

import (
	healthResponse "rea/porticos/internal/modules/health/domain/dto/responses"
)

// Logica de negocio, aquí se invocan los repository que se integran con bases de datos
// o llamados a otras APIS
type HealthUseCase struct{}

func NewHealthUseCase() *HealthUseCase {
	return &HealthUseCase{}
}

func (h *HealthUseCase) CheckHealth() (*healthResponse.HealthResponse, error) {

	// Aqui podriamos agregar otras validaciones:
	// - Verificar conexión a DB
	// - Verificación con servicios externos
	// Si alguno falla, se retorna error HTTP 503
	return healthResponse.NewHealthResponse(), nil
}
