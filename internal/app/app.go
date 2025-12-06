package app

import (
	"fmt"
	"net/http"
	"strconv"

	carHandlers "github.com/A1exanderShin/autoglobal/internal/cars/handlers"
	"github.com/A1exanderShin/autoglobal/internal/cars/repository"
	"github.com/A1exanderShin/autoglobal/internal/cars/service"
	"github.com/A1exanderShin/autoglobal/internal/config"
	httpHandlers "github.com/A1exanderShin/autoglobal/internal/http/handlers"
	"github.com/A1exanderShin/autoglobal/internal/http/middleware"
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

	// 2. Прогон миграций (инициализация структуры БД)
	if err := storage.RunMigrations(pool); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	// -----------------------------
	// Cars module (репозиторий → сервис → хендлеры)
	// -----------------------------

	// Репозиторий — работает с PostgreSQL
	carRepo := repository.NewPostgresCarRepository(pool)

	// Service — бизнес-логика (валидация, обработка ошибок, управление сущностями)
	carSvc := service.New(carRepo)

	// Handlers — HTTP-уровень (получают запросы, вызывают сервис, формируют ответы)
	carHandlers := carHandlers.NewCarHandlers(carSvc)

	// 3. Создание HTTP-роутера
	router := chi.NewRouter()

	// 4. Middleware (добавляют функциональность на уровне HTTP)
	router.Use(middleware.RequestID) // каждому запросу присваивается уникальный ID
	router.Use(middleware.Logger)    // логирование запросов/ответов
	router.Use(middleware.Recoverer) // защита от паник — возвращает JSON 500

	// Health-check для автоматизированных систем мониторинга
	router.Get("/health", httpHandlers.Health)

	// Cars API — REST эндпоинты
	router.Route("/cars", func(r chi.Router) {
		r.Post("/", carHandlers.CreateCar) // создание машины
		r.Get("/{id}", carHandlers.GetCar) // получение машины по ID
		r.Get("/", carHandlers.ListCars)   // список всех машин
		r.Put("/{id}", carHandlers.UpdateCar)
		r.Delete("/{id}", carHandlers.DeleteCar)
	})

	// 5. Создание объекта приложения (чтобы можно было расширять в будущем)
	app := &App{
		Cfg:    *cfg,
		DB:     pool,
		Router: router,
	}

	// 6. Создание и запуск HTTP-сервера
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.HTTP.Port), // порт из конфига
		Handler: app.Router,                        // корневой роутер
	}

	// Запускаем сервис (блокирует выполнение)
	return srv.ListenAndServe()
}
