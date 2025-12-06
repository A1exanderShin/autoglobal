package repository

import (
	"context"

	"github.com/A1exanderShin/autoglobal/internal/cars"
)

type CarRepository interface {
	Create(ctx context.Context, c cars.Car) (int64, error)
	GetByID(ctx context.Context, id int64) (*cars.Car, error)
	List(ctx context.Context) ([]cars.Car, error)
	Update(ctx context.Context, id int64, car cars.Car) error
	Delete(ctx context.Context, id int64) error
}
