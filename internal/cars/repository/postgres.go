package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/A1exanderShin/autoglobal/internal/cars"
	"github.com/A1exanderShin/autoglobal/internal/cars/dto"
	"github.com/jackc/pgx/v5"
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
func (r *PostgresCarRepository) Create(
	ctx context.Context,
	c cars.Car,
	userID *int64,
) (int64, error) {

	query := `
		INSERT INTO cars (brand, model, year, price, url, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
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
		c.URL,
		userID,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByID — вернуть одну запись машины по id
// Если записи нет, вернётся pgx.ErrNoRows (это обработает сервисный слой)
func (r *PostgresCarRepository) GetByID(ctx context.Context, id int64) (*CarRow, error) {
	query := `
		SELECT id, brand, model, year, price, url, user_id
		FROM cars
		WHERE id = $1
	`

	var c CarRow
	err := r.db.QueryRow(ctx, query, id).
		Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.Price, &c.URL, &c.UserID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &c, nil
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

func (r *PostgresCarRepository) ListFiltered(ctx context.Context, f dto.CarFilters) ([]cars.Car, error) {
	query := "SELECT id, brand, model, year, price FROM cars"

	var where []string
	var args []interface{}
	argID := 1

	// Здесь мы скоро добавим условия where = append(...)
	if f.Brand != "" {
		where = append(where, fmt.Sprintf("brand ILIKE $%d", argID))
		args = append(args, "%"+f.Brand+"%")
		argID++
	}

	if f.Model != "" {
		where = append(where, fmt.Sprintf("model ILIKE $%d", argID))
		args = append(args, "%"+f.Model+"%")
		argID++
	}

	if f.MinYear > 0 {
		where = append(where, fmt.Sprintf("year >= $%d", argID))
		args = append(args, f.MinYear)
		argID++
	}

	if f.MaxYear > 0 {
		where = append(where, fmt.Sprintf("year <= $%d", argID))
		args = append(args, f.MaxYear)
		argID++
	}

	if f.MinPrice > 0 {
		where = append(where, fmt.Sprintf("price >= $%d", argID))
		args = append(args, f.MinPrice)
		argID++
	}

	if f.MaxPrice > 0 {
		where = append(where, fmt.Sprintf("price <= $%d", argID))
		args = append(args, f.MaxPrice)
		argID++
	}

	// Добавляем WHERE если есть условия
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	switch f.Sort {
	case "price_asc":
		query += " ORDER BY price ASC"
	case "price_desc":
		query += " ORDER BY price DESC"
	case "year_asc":
		query += " ORDER BY year ASC"
	case "year_desc":
		query += " ORDER BY year DESC"
	default:
		query += " ORDER BY id DESC"
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}

	page := f.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []cars.Car{}

	for rows.Next() {
		var car cars.Car
		if err := rows.Scan(&car.ID, &car.Brand, &car.Model, &car.Year, &car.Price); err != nil {
			return nil, err
		}
		list = append(list, car)
	}

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

func (r *PostgresCarRepository) ExistsByURL(ctx context.Context, url string) (bool, error) {
	const query = `SELECT 1 FROM cars WHERE url = $1 LIMIT 1`

	var tmp int
	err := r.db.QueryRow(ctx, query, url).Scan(&tmp)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
