package service

import (
	"context"
	"errors"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/A1exanderShin/autoglobal/internal/cars/repository"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("car not found")
)

type CarRepository interface {
	Create(ctx context.Context, c cars.Car) (int64, error)
	GetByID(ctx context.Context, id int64) (*cars.Car, error)
	ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error)
	Update(ctx context.Context, id int64, car cars.Car) error
	Delete(ctx context.Context, id int64) error
	ExistsByURL(ctx context.Context, url string) (bool, error)
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
		URL:   req.URL,
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

func (s *Service) ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error) {
	// 1. Валидации диапазонов
	if f.MinYear != 0 && f.MaxYear != 0 && f.MinYear > f.MaxYear {
		return nil, ErrValidation
	}

	if f.MinPrice != 0 && f.MaxPrice != 0 && f.MinPrice > f.MaxPrice {
		return nil, ErrValidation
	}

	// 2. Контекст с таймаутом
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 3. Вызов репозитория
	list, err := s.repo.ListFiltered(cctx, f)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) UpdateCar(ctx context.Context, id int64, req dto.UpdateCarRequest) error {
	currentYear := time.Now().Year()

	// Валидация входных данных
	if req.Brand == "" || req.Model == "" {
		return ErrValidation
	}
	if req.Year < 1900 || req.Year > currentYear {
		return ErrValidation
	}
	if req.Price < 0 {
		return ErrValidation
	}

	// Подготовка контекста с таймаутом
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Формируем доменную сущность
	car := cars.Car{
		ID:    id,
		Brand: req.Brand,
		Model: req.Model,
		Year:  req.Year,
		Price: req.Price,
		URL:   req.URL,
	}

	// Вызываем репозиторий
	err := s.repo.Update(cctx, id, car)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Service) DeleteCar(ctx context.Context, id int64) error {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.repo.Delete(cctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *Service) ExistsByURL(ctx context.Context, url string) (bool, error) {
	return s.repo.ExistsByURL(ctx, url)
}
