package main

import (
	"AirPort/internal/config"
	"AirPort/internal/handlers/board"
	"AirPort/internal/handlers/tickets"
	"AirPort/internal/handlers/user"
	control "AirPort/internal/handlers/userControl"
	"AirPort/package/database"
	"AirPort/package/server"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загрузка основного конфига
	var cfg config.ServerConf
	if err := cfg.ReadConfig(); err != nil {
		log.Fatalf("Ошибка чтения основного конфига: %s", err)
	}

	// Загрузка конфига базы данных
	var dbConf config.StorageConfig
	if err := dbConf.ReadConfig(); err != nil {
		log.Fatalf("Ошибка чтения основного конфига: %s", err)
	}
	// Подключение к БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.OpenDBClient(ctx, dbConf)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer pool.Close()

	// Инициализация gin
	router := gin.Default()

	// Middleware для CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Инициализация роутов
	// -- для User
	userHandler := user.NewHandler(pool)
	userHandler.RegisterHandler(router)

	// -- для Board
	boardHandler := board.NewHandler(pool)
	boardHandler.RegisterHandler(router)

	// -- для Control
	controlHandler := control.NewHandler(pool)
	controlHandler.RegisterHandler(router)

	// -- для Tickets
	ticketsHandler := tickets.NewHandler(pool)
	ticketsHandler.RegisterHandler(router)

	// Запуск сервера
	server := &server.Server{}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.RunServer(cfg.Port, router); err != nil {
			log.Fatalf("Ошибка при запуске сервера: %s", err)
		}
	}()

	log.Printf("\033[32mСервер запущен на: %s:%s\n\033[0m", cfg.Host, cfg.Port)

	<-done

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.StopServer(ctxShutdown); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %v", err)
	}
	log.Println("Сервер успешно остановлен")
}
