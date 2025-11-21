package app

import (
	"net/http"
	"strconv"

	"github.com/A1exanderShin/autoglobal/internal/config"
	"github.com/A1exanderShin/autoglobal/internal/http/handlers"
	"github.com/A1exanderShin/autoglobal/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App — главный объект приложения
// Хранит конфигурацию, подключение к БД и HTTP-роутер
type App struct {
	Cfg    config.Config
	DB     *pgxpool.Pool
	Router *chi.Mux
}

// Run — точка входа для сервиса
// Создаёт подключение к PostgreSQL, инициализирует роутер и запускает HTTP-сервер
func Run(cfg *config.Config) error {
	// 1. Инициализация пула PostgreSQL
	// Подключаемся к базе и проверяем соединение через Ping()
	pool, err := storage.NewPostgresPool(cfg.Postgres)
	if err != nil {
		return err
	}

	// 2. Создаём новый HTTP-роутер на базе chi
	router := chi.NewRouter()

	router.Get("/health", handlers.Health)

	// 3. Сборка объекта App
	// Сохраняем конфиг, базу и роутер в структуре
	app := &App{
		Cfg:    *cfg,
		DB:     pool,
		Router: router,
	}

	// 4. Создаём HTTP-сервер
	// Addr — порт, который сервис будет слушать
	// Handler — корневой роутер chi
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler: app.Router,
	}

	// 5. Запускаем HTTP-сервер
	// Блокирующая операция: пока не упадёт — выполнение main.go стоит
	return srv.ListenAndServe()
}
