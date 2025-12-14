package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/users"
	"github.com/A1exanderShin/autoglobal/internal/users/dto"
	"github.com/A1exanderShin/autoglobal/internal/users/repository"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	repo   repository.UserRepository
	jwtKey []byte
}

func NewUsersService(repo repository.UserRepository, jwtKey []byte) *UsersService {
	return &UsersService{repo: repo, jwtKey: jwtKey}
}

func (s *UsersService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Проверяем, существует ли пользователь
	_, err := s.repo.GetByEmail(ctx, email)

	switch {
	case err == nil:
		// пользователь найден
		return nil, fmt.Errorf("email already exists")

	case err == repository.ErrNotFound:
		// пользователя нет — можно регистрировать
		// продолжаем

	case err != nil:
		// SQL ошибка или другая
		return nil, err
	}

	// Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаём доменную модель
	u := users.User{
		Email:    email,
		Password: string(hash),
	}

	// Пишем в базу
	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(id, email)

	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := s.generateRefreshToken()

	if err != nil {
		return nil, err
	}

	err = s.repo.SaveRefreshToken(ctx, id, refreshToken, expiresAt)

	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		UserID:       id,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (s *UsersService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	user, err := s.repo.GetByEmail(ctx, email)

	if err == repository.ErrNotFound {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)

	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := s.generateRefreshToken()

	if err != nil {
		return nil, err
	}

	err = s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt)

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (s *UsersService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err == repository.ErrNotFound {
		return nil, fmt.Errorf("unauthorized")
	}
	if err != nil {
		return nil, err
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, refreshToken)
		return nil, fmt.Errorf("unauthorized")
	}

	// удаляем старый refresh token (rotation)
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(rt.UserID, rt.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiresAt, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveRefreshToken(ctx, rt.UserID, newRefreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *UsersService) Logout(ctx context.Context, refreshToken string) error {
	token := strings.TrimSpace(refreshToken)
	if token == "" {
		return nil
	}

	err := s.repo.DeleteRefreshToken(ctx, token)
	if err == repository.ErrNotFound {
		return nil
	}

	return err
}

func (s *UsersService) generateAccessToken(userID int64, email string) (string, error) {
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtKey)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *UsersService) generateRefreshToken() (string, time.Time, error) {
	const tokenSize = 32

	b := make([]byte, tokenSize)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	return token, expiresAt, nil
}
