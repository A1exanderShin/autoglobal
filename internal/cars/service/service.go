package service

import (
	"context"
	"errors"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/A1exanderShin/autoglobal/internal/cars/repository"
	"github.com/A1exanderShin/autoglobal/internal/http/middleware"
)

var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("car not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Service struct {
	repo repository.CarRepository
}

func New(repo repository.CarRepository) *Service {
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

	// userID живёт только в рамках запроса
	var userID *int64
	if id, ok := middleware.UserIDFromContext(cctx); ok {
		userID = &id
	}

	id, err := s.repo.Create(cctx, car, userID)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) GetCar(ctx context.Context, id int64) (*cars.Car, error) {
	carRow, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &cars.Car{
		ID:    carRow.ID,
		Brand: carRow.Brand,
		Model: carRow.Model,
		Year:  carRow.Year,
		Price: carRow.Price,
		URL:   carRow.URL,
	}, nil
}

func (s *Service) ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error) {
	if f.MinYear != 0 && f.MaxYear != 0 && f.MinYear > f.MaxYear {
		return nil, ErrValidation
	}
	if f.MinPrice != 0 && f.MaxPrice != 0 && f.MinPrice > f.MaxPrice {
		return nil, ErrValidation
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	list, err := s.repo.ListFiltered(cctx, f)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) UpdateCar(ctx context.Context, id int64, req dto.UpdateCarRequest) error {
	currentYear := time.Now().Year()

	if req.Brand == "" || req.Model == "" {
		return ErrValidation
	}
	if req.Year < 1900 || req.Year > currentYear {
		return ErrValidation
	}
	if req.Price < 0 {
		return ErrValidation
	}

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	carRow, err := s.repo.GetByID(cctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	if carRow.UserID == nil || *carRow.UserID != userID {
		return ErrForbidden
	}

	car := cars.Car{
		ID:    id,
		Brand: req.Brand,
		Model: req.Model,
		Year:  req.Year,
		Price: req.Price,
		URL:   req.URL,
	}

	return s.repo.Update(cctx, id, car)
}

func (s *Service) DeleteCar(ctx context.Context, id int64) error {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	carRow, err := s.repo.GetByID(cctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	if carRow.UserID == nil || *carRow.UserID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(cctx, id)
}

func (s *Service) ExistsByURL(ctx context.Context, url string) (bool, error) {
	return s.repo.ExistsByURL(ctx, url)
}
