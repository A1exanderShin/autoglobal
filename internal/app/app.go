package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	carHandlers "github.com/A1exanderShin/autoglobal/internal/cars/handlers"
	"github.com/A1exanderShin/autoglobal/internal/cars/repository"
	"github.com/A1exanderShin/autoglobal/internal/cars/service"
	"github.com/A1exanderShin/autoglobal/internal/config"
	httpHandlers "github.com/A1exanderShin/autoglobal/internal/http/handlers"
	"github.com/A1exanderShin/autoglobal/internal/http/middleware"
	"github.com/A1exanderShin/autoglobal/internal/parser"
	"github.com/A1exanderShin/autoglobal/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App — главный объект приложения.
// Хранит конфигурацию, подключение к БД и HTTP-роутер.
type App struct {
	Cfg    config.Config
	DB     *pgxpool.Pool
	Router *chi.Mux
}

// Run — точка входа для сервиса.
// Последовательность:
// 1) Подключение к БД
// 2) Прогон миграций
// 3) Инициализация роутера и middleware
// 4) Подключение модулей (репозиторий → сервис → хендлеры)
// 5) Запуск HTTP-сервера
func Run(cfg *config.Config) error {

	// 1. Подключение к PostgreSQL
	pool, err := storage.NewPostgresPool(cfg.Postgres)
	if err != nil {
		return err
	}

	// 2. Прогон миграций
	if err := storage.RunMigrations(pool); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	// 3. Инициализация модулей Cars
	carRepo := repository.NewPostgresCarRepository(pool)
	carSvc := service.New(carRepo)

	// --- временный запуск парсера ---
	p := parser.NewParser(carSvc)
	go p.ParseAll(context.Background(), "https://auto.drom.ru/", 100)
	// ---------------------------------

	// 4. Роутер + middleware
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", httpHandlers.Health)

	carHandlers := carHandlers.NewCarHandlers(carSvc)
	router.Route("/cars", func(r chi.Router) {
		r.Post("/", carHandlers.CreateCar)
		r.Get("/{id}", carHandlers.GetCar)
		r.Get("/search", carHandlers.ListFiltered)
		r.Put("/{id}", carHandlers.UpdateCar)
		r.Delete("/{id}", carHandlers.DeleteCar)
	})

	// 5. Запуск HTTP сервера
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler: router,
	}

	return srv.ListenAndServe()
}
