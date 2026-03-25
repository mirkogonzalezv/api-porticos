package repository

import (
	"context"
	"time"

	"rea/porticos/internal/modules/pasos/domain/entities"
)

type ListPasosFilter struct {
	From       time.Time
	To         time.Time
	VehiculoID string
	PorticoID  string
	Limit      int
	Offset     int
}

type SummaryPasosFilter struct {
	From       time.Time
	To         time.Time
	VehiculoID string
	PorticoID  string
	GroupBy    string
}

type PasoPorticoRepository interface {
	Create(ctx context.Context, paso *entities.PasoPortico) (*entities.PasoPortico, error)
	CreateBatch(ctx context.Context, pasos []*entities.PasoPortico) ([]entities.PasoPortico, error)
	CreateCapture(ctx context.Context, paso *entities.PasoCapturado) (*entities.PasoCapturado, error)
	GetByID(ctx context.Context, ownerID, id string) (*entities.PasoPortico, error)
	ListByOwnerRange(ctx context.Context, ownerID string, filter ListPasosFilter) ([]entities.PasoPortico, error)
	ListAllRange(ctx context.Context, filter ListPasosFilter) ([]entities.PasoPortico, error)
	ListCapturadosByOwnerRange(ctx context.Context, ownerID string, filter ListPasosFilter) ([]entities.PasoCapturado, error)
	ListCapturadosAllRange(ctx context.Context, filter ListPasosFilter) ([]entities.PasoCapturado, error)
	SummaryByOwnerRange(ctx context.Context, ownerID string, filter SummaryPasosFilter) ([]entities.ResumenPeriodo, error)
}
