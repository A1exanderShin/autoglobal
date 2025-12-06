package repository

import (
	"context"
	"errors"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("car not found")

// PostgresCarRepository — реализация CarRepository,
// которая работает с PostgreSQL через pgxpool
type PostgresCarRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCarRepository — конструктор репозитория
// Получает пул соединений и возвращает экземпляр репозитория
func NewPostgresCarRepository(db *pgxpool.Pool) *PostgresCarRepository {
	return &PostgresCarRepository{db: db}
}

// Create — создаёт новую запись автомобиля в таблице cars
// Принимает доменную модель cars.Car (без ID) и возвращает сгенерированный ID
// SQL: INSERT INTO ... RETURNING id
func (r *PostgresCarRepository) Create(ctx context.Context, c cars.Car) (int64, error) {
	query := `
		INSERT INTO cars (brand, model, year, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64

	// QueryRow — выполнение SQL-скрипта для одиночной строки
	// Scan — считывает значение RETURNING id в переменную id
	err := r.db.QueryRow(ctx, query,
		c.Brand,
		c.Model,
		c.Year,
		c.Price,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByID — вернуть одну запись машины по id
// Если записи нет, вернётся pgx.ErrNoRows (это обработает сервисный слой)
func (r *PostgresCarRepository) GetByID(ctx context.Context, id int64) (*cars.Car, error) {
	query := `
		SELECT id, brand, model, year, price
		FROM cars
		WHERE id = $1
	`

	// QueryRow — получаем ровно одну строку
	row := r.db.QueryRow(ctx, query, id)

	var car cars.Car

	// Scan мапит значения колонок в структуру Car
	err := row.Scan(
		&car.ID,
		&car.Brand,
		&car.Model,
		&car.Year,
		&car.Price,
	)
	if err != nil {
		return nil, err
	}

	return &car, nil
}

// List — возвращает список всех машин
// SELECT ... FROM cars
func (r *PostgresCarRepository) List(ctx context.Context) ([]cars.Car, error) {
	query := `
		SELECT id, brand, model, year, price
		FROM cars
	`

	// Query — делаем запрос, который возвращает много строк
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	// ВАЖНО: обязательно закрываем cursor, чтобы не было утечек
	defer rows.Close()

	list := []cars.Car{}

	// Проходим по всем строкам результата.
	for rows.Next() {
		var car cars.Car
		// Scan читает текущую строку в car.
		err := rows.Scan(
			&car.ID,
			&car.Brand,
			&car.Model,
			&car.Year,
			&car.Price,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, car)
	}

	// Возвращаем собранный список.
	return list, nil
}

func (r *PostgresCarRepository) Update(ctx context.Context, id int64, c cars.Car) error {
	query := `
		UPDATE cars
		SET brand = $1, model = $2, year = $3, price = $4
		WHERE id = $5
	`
	res, err := r.db.Exec(ctx, query, c.Brand, c.Model, c.Year, c.Price, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresCarRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM cars WHERE id = $1`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
