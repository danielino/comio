package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/danielino/comio/internal/config"
	"github.com/danielino/comio/internal/monitoring"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	srv    *http.Server
	cfg    *config.Config
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	router := gin.New()
	
	return &Server{
		router: router,
		cfg:    cfg,
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  parseDuration(s.cfg.Server.ReadTimeout),
		WriteTimeout: parseDuration(s.cfg.Server.WriteTimeout),
	}

	monitoring.Log.Info("Starting server", zap.String("addr", s.srv.Addr))

	if s.cfg.Server.TLS.Enabled {
		return s.srv.ListenAndServeTLS(s.cfg.Server.TLS.CertFile, s.cfg.Server.TLS.KeyFile)
	}
	return s.srv.ListenAndServe()
}

// Stop stops the server gracefully
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func parseDuration(d string) time.Duration {
	dur, err := time.ParseDuration(d)
	if err != nil {
		return 30 * time.Second
	}
	return dur
}
