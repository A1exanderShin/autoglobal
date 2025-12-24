package repository

import (
	"context"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
)

type CarRepository interface {
	Create(ctx context.Context, car cars.Car, userID *int64) (int64, error)
	GetByID(ctx context.Context, id int64) (*CarRow, error)
	ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error)
	Update(ctx context.Context, id int64, car cars.Car) error
	Delete(ctx context.Context, id int64) error
	ExistsByURL(ctx context.Context, url string) (bool, error)
}
