package repository

import (
	"context"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
)

type CarRepository interface {
	Create(ctx context.Context, c cars.Car) (int64, error)
	GetByID(ctx context.Context, id int64) (*cars.Car, error)
	ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error) // новый
	Update(ctx context.Context, id int64, car cars.Car) error
	Delete(ctx context.Context, id int64) error
	ExistsByURL(ctx context.Context, url string) (bool, error)
}
