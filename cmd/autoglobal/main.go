package main

import (
	"fmt"
	"os"

	"github.com/A1exanderShin/autoglobal/internal/app"
	"github.com/A1exanderShin/autoglobal/internal/config"
)

func main() {
	// 1. Получаем путь к конфигу из переменной окружения CONFIG_PATH.
	// Если переменная пустая — config.Load подставит значение по умолчанию.
	path := os.Getenv("CONFIG_PATH")

	// 2. Загружаем конфигурацию приложения из YAML.
	// Load:
	//  - считывает YAML-файл,
	//  - мапит поля в структуры,
	//  - валидирует данные (если указаны теги),
	//  - возвращает *Config или ошибку.
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Printf("load config err: %v\n", err)
		os.Exit(1)
	}

	// 3. Запускаем приложение.
	// Run:
	//  - подключается к базе PostgreSQL,
	//  - инициализирует роутер,
	//  - создаёт HTTP-сервер,
	//  - начинает слушать порт.
	err = app.Run(cfg)
	if err != nil {
		fmt.Printf("run err: %v\n", err)
		os.Exit(1)
	}
}
