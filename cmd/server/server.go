package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  1 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	log.Info().Msgf("running on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
