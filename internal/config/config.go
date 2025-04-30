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
		log.Printf("Ошибка при чтении env файла: %s", err)
		return err
	}

	return nil
}
