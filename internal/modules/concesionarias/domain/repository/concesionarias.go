package repository

import (
	"context"

	"rea/porticos/internal/modules/concesionarias/domain/entities"
)

type ListConcesionariasFilter struct {
	Limit  int
	Offset int
	Estado string
}

type ConcesionariaRepository interface {
	Create(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error)
	List(ctx context.Context, filter ListConcesionariasFilter) ([]entities.Concesionaria, error)
	GetByID(ctx context.Context, id string) (*entities.Concesionaria, error)
	Update(ctx context.Context, concesionaria *entities.Concesionaria) (*entities.Concesionaria, error)
	Delete(ctx context.Context, id string) error
}
