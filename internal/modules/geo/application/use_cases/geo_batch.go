package usecases

import (
	"context"

	"rea/porticos/internal/modules/geo/domain/dtos/requests"
	domainErrors "rea/porticos/pkg/errors"
)

type GeoBatchUseCase struct{}

func NewGeoBatchUseCase() *GeoBatchUseCase {
	return &GeoBatchUseCase{}
}

func (uc *GeoBatchUseCase) ValidateBatch(ctx context.Context, req *requests.GeoBatchRequest) error {
	_ = ctx
	if req == nil {
		return domainErrors.NewValidationError("GEO_BATCH_REQUIRED", "batch es obligatorio")
	}
	return req.Validate()
}
