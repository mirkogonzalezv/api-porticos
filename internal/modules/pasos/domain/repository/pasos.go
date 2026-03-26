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

type PasoPorticoRepository interface {
	Create(ctx context.Context, paso *entities.PasoPortico) (*entities.PasoPortico, error)
	CreateBatch(ctx context.Context, pasos []*entities.PasoPortico) ([]entities.PasoPortico, error)
	CreateCapture(ctx context.Context, paso *entities.PasoCapturado) (*entities.PasoCapturado, error)
	CreateCapturesBatch(ctx context.Context, pasos []*entities.PasoCapturado) error
	CreateConfirmadosBatch(ctx context.Context, pasos []*entities.PasoPortico) error
	AcquireIdempotencyKey(ctx context.Context, ownerID, key, scope string, ttl time.Duration) (bool, error)
	GetByID(ctx context.Context, ownerID, id string) (*entities.PasoPortico, error)
	ListByOwnerRange(ctx context.Context, ownerID string, filter ListPasosFilter) ([]entities.PasoPortico, error)
	ListAllRange(ctx context.Context, filter ListPasosFilter) ([]entities.PasoPortico, error)
	ListCapturadosByOwnerRange(ctx context.Context, ownerID string, filter ListPasosFilter) ([]entities.PasoCapturado, error)
	ListCapturadosAllRange(ctx context.Context, filter ListPasosFilter) ([]entities.PasoCapturado, error)
}
