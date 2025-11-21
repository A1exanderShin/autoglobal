package config

import "github.com/ilyakaznacheev/cleanenv"

type HTTPConfig struct {
	Port int `yaml:"port"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}
type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// Прочитать config, распарсить в структуры, вернуть ошибку (если кривой конфиг)
func Load(path string) (*Config, error) {
	if path == "" {
		path = "./config/local.yaml"
	}

	cfg := Config{}
	// Открыть YAML-файл, читать этот файл, превратить YAML в структуры,
	// Заполнить значения из YAML, валидация по тегам
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
