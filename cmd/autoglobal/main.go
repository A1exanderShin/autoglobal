package main

import (
	"fmt"
	"github.com/A1exanderShin/autoglobal/internal/config"
	"os"
)

func main() {
	// Получить переменную окружения от ОС
	path := os.Getenv("CONFIG_PATH")

	// Вызываешь свою функцию, которая:
	// определяет путь к конфигу
	// открывает YAML файл
	// парсит YAML → структуры
	// валидирует
	// возвращает или ошибку, или готовый Config
	cfg, err := config.Load(path)

	// Проверяешь ошибку загрузки конфига
	if err != nil {
		fmt.Printf("load config err: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("config loaded:", cfg.HTTP.Port)
}
