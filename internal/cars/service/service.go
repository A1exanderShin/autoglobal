package service

import (
	"context"
	"errors"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("car not found")
)

type CarRepository interface {
	Create(ctx context.Context, c cars.Car) (int64, error)
	GetByID(ctx context.Context, id int64) (*cars.Car, error)
	List(ctx context.Context) ([]cars.Car, error)
}

type Service struct {
	repo CarRepository
}

func New(repo CarRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCar(ctx context.Context, req dto.CreateCarRequest) (int64, error) {
	currentYear := time.Now().Year()

	if req.Brand == "" || req.Model == "" {
		return 0, ErrValidation
	}
	if req.Year < 1900 || req.Year > currentYear {
		return 0, ErrValidation
	}
	if req.Price < 0 {
		return 0, ErrValidation
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	car := cars.Car{
		Brand: req.Brand,
		Model: req.Model,
		Year:  req.Year,
		Price: req.Price,
	}

	id, err := s.repo.Create(cctx, car)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) GetCar(ctx context.Context, id int64) (*cars.Car, error) {
	car, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return car, nil
}

func (s *Service) ListCars(ctx context.Context) ([]cars.Car, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil

}
