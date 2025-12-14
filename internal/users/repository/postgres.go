package repository

import (
	"context"
	"errors"
	"time"

	"github.com/A1exanderShin/autoglobal/internal/users"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUsersRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUsersRepository(db *pgxpool.Pool) *PostgresUsersRepository {
	return &PostgresUsersRepository{db: db}
}

func (r *PostgresUsersRepository) CreateUser(ctx context.Context, u users.User) (int64, error) {
	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, u.Email, u.Password).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, ErrAlreadyExists
			}
		}
		return 0, err
	}

	return id, nil
}

func (r *PostgresUsersRepository) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	var u users.User

	query := `
		SELECT id, email, password
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &u, nil

}

func (r *PostgresUsersRepository) GetUserByID(ctx context.Context, id int64) (*users.User, error) {
	var u users.User

	query := `
		SELECT id, email, password
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *PostgresUsersRepository) SaveRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, refreshToken, expiresAt)
	return err

}

func (r *PostgresUsersRepository) GetRefreshToken(ctx context.Context, refreshToken string) (*RefreshToken, error) {
	var token RefreshToken

	query := `
		SELECT user_id, token, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`

	err := r.db.QueryRow(ctx, query, token).Scan(&token.UserID, &token.Token, &token.ExpiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &token, nil

}

func (r *PostgresUsersRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := r.db.Exec(ctx, query, token)

	return err
}
