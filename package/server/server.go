package server

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) RunServer(port string, h http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		Handler:        h,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) StopServer(c context.Context) error {
	return s.httpServer.Shutdown(c)
}
