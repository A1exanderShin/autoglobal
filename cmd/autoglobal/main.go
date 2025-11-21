package main

import (
	"fmt"
	"github.com/AlexanderShin/autoglobal/internal/config"
	"os"
)

func main() {
	path := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Printf("load config err: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("config loaded:", cfg.HTTP.Port)
}
