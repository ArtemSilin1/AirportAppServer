package main

import (
	"AirPort/internal/config"
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

	// Инициализация gin
	router := gin.Default()

	// Запуск сервера
	server := &server.Server{}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.RunServer(cfg.Port, router); err != nil {
			log.Fatalf("Ошибка при запуске сервера: %s", err)
		}
	}()

	log.Printf("Сервер запущен на: %s %s\n", cfg.Host, cfg.Port)

	<-done

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.StopServer(ctxShutdown); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %v", err)
	}
	log.Println("Сервер успешно остановлен")
}
