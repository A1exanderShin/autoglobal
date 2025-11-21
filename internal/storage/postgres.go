package storage

import (
	"context"
	"fmt"

	"github.com/A1exanderShin/autoglobal/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Создаёт пул соединений к PostgreSQL и возвращает его
// Используется в app.go, чтобы сервис мог выполнять SQL-операции
func NewPostgresPool(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	// DSN (Data Source Name) — это обычная строка, которая описывает:
	// логин
	// пароль
	// хост
	// порт
	// имя базы
	// То есть куда и как Go должен подключиться к PostgreSQL
	// В Go драйвер pgx принимает DSN в виде строки
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)

	// Создание контекста
	ctx := context.Background()

	// Создать пул подключений
	// Что происходит:
	// Запускается логика подключения к PostgreSQL
	// Создаётся пул соединений (например 4–8 подключений)
	// Если что-то не так (контейнер не запущен) → ошибка
	// Пул — это объект, через который сервис будет выполнять запросы
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	// Что это делает:
	// отправляет тестовый запрос в PostgreSQL
	// проверяет, что база действительно отвечает
	// выявляет неправильные пароли, DSN, порты
	// защищает от ситуации «пул создался, но база недоступна»

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
