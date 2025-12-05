package storage

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // обязательно! goose использует database/sql
	"github.com/pressly/goose/v3"
)

func RunMigrations(pool *pgxpool.Pool) error {
	// Получаем DSN из pgxpool
	dsn := pool.Config().ConnConfig.ConnString()

	// Открываем обычное подключение к базе через database/sql
	// Goose работает ТОЛЬКО с *sql.DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sql DB for migrations: %w", err)
	}
	defer db.Close()

	// Отключаем лишний логгинг Goose
	goose.SetLogger(goose.NopLogger())
	goose.SetDialect("postgres")

	// Запускаем миграции
	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("goose migration up failed: %w", err)
	}

	return nil
}
