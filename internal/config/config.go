package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServerConf struct {
	Port string `env:"PORT" env-default:"8081"`
	Host string `env:"Host" env-default:"localhost"`
}

func (s *ServerConf) ReadConfig() error {
	err := cleanenv.ReadConfig("internal/config/.env", s)
	if err != nil {
		log.Printf("Ошибка при чтении файла с конфигом: %s", err)
		return err
	}

	return nil
}

type StorageConfig struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	Database string `env:"DB_DATABASE" env-default:"postgres"`
	Username string `env:"DB_USERNAME" env-default:"postgres"`
	Password string `env:"DB_PASSWORD"`
}

func (c *StorageConfig) ReadConfig() error {
	err := cleanenv.ReadConfig("internal/config/.env", c)
	if err != nil {
		log.Printf("Ошибка при чтении файла с конфигом: %s", err)
		return err
	}

	return nil
}
