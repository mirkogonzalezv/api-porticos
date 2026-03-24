package repository

import (
	"context"

	"rea/porticos/internal/modules/porticos/domain/entities"
)

type ListPorticosFilter struct {
	Limit  int
	Offset int
}

type PorticoRepository interface {
	Create(ctx context.Context, portico *entities.Portico) (*entities.Portico, error)
	List(ctx context.Context, filter ListPorticosFilter) ([]entities.Portico, error)
	GetByID(ctx context.Context, id string) (*entities.Portico, error)
	GetByCodigo(ctx context.Context, codigo string) (*entities.Portico, error)
	ListNearby(ctx context.Context, lat, lng, maxDistanceMeters float64) ([]entities.Portico, error)
	FindByTrajectory(ctx context.Context, lineWKT string) ([]entities.Portico, error)
	Update(ctx context.Context, portico *entities.Portico) (*entities.Portico, error)
	Delete(ctx context.Context, id string) error
}
