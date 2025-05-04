package database

import (
	"AirPort/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func OpenDBClient(ctx context.Context, config config.StorageConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при подключении к базе данных: %s", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Попытка соединения не удалась: %s", err)
	}

	log.Println("\033[32mПодключено к бд\033[0m")
	return pool, nil
}
