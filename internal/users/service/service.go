package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/A1exanderShin/autoglobal/internal/users"
	"github.com/A1exanderShin/autoglobal/internal/users/dto"
	"github.com/A1exanderShin/autoglobal/internal/users/repository"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	repo   repository.UsersRepository
	jwtKey []byte
}

func NewUsersService(repo repository.UsersRepository, jwtKey []byte) *UsersService {
	return &UsersService{repo: repo, jwtKey: jwtKey}
}

func (s *UsersService) Register(ctx context.Context, req dto.RegisterRequest) (int64, error) {
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if email == "" {
		return 0, fmt.Errorf("email is required")
	}
	if password == "" {
		return 0, fmt.Errorf("password is required")
	}

	// Проверяем, существует ли пользователь
	_, err := s.repo.GetByEmail(ctx, email)

	switch {
	case err == nil:
		// пользователь найден
		return 0, fmt.Errorf("email already exists")

	case err == repository.ErrNotFound:
		// пользователя нет — можно регистрировать
		// продолжаем

	case err != nil:
		// SQL ошибка или другая
		return 0, err
	}

	// Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Создаём доменную модель
	u := users.User{
		Email:    email,
		Password: string(hash),
	}

	// Пишем в базу
	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *UsersService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	return nil, nil
}

func (s *UsersService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	return nil, nil
}

func (s *UsersService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}
