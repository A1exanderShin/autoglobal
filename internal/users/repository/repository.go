package repository

import (
	"context"
	"errors"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/users"
)

var ErrNotFound = errors.New("users not found")
var ErrAlreadyExists = errors.New("user already exists")

type UserRepository interface {
	CreateUser(ctx context.Context, u users.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*users.User, error)
	GetUserByID(ctx context.Context, id int64) (*users.User, error)
	SaveRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, refreshToken string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
