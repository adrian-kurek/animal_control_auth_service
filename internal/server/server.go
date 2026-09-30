package server

import (
	"context"
	"net/http"
	"time"

	"github.com/adrian-kurek/animal_control_auth_service/internal/auth"
)

const (
	defaultReadTimeout  = 50 * time.Second
	defaultWriteTimeout = 50 * time.Second
	defaultIdleTimeout  = 30 * time.Second
)

type DependencyConfig struct {
	port        string
	authHandler auth.Handler
}

func NewDependencyConfig(
	port string,
	authHandler auth.Handler,
) *DependencyConfig {
	return &DependencyConfig{
		port:        port,
		authHandler: authHandler,
	}
}

type HTTP struct {
	config *DependencyConfig
	server *http.Server
	router *http.ServeMux
}

func NewHTTP(config *DependencyConfig) *HTTP {
	return &HTTP{
		config: config,
		router: http.NewServeMux(),
	}
}

func (s *HTTP) Start() error {
	s.SetupRoutes()
	s.server = &http.Server{
		Addr:         ":" + s.config.port,
		Handler:      s.router,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}
	return s.server.ListenAndServe()
}

func (s *HTTP) SetupRoutes() {
	authRoute := auth.NewRoute(&s.config.authHandler)
	authRoute.Setup(s.router)
}

func (s *HTTP) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
