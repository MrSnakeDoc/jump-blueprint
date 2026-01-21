// internal/httpserver/server.go
package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/MrSnakeDoc/jump/internal/config"
	"github.com/MrSnakeDoc/jump/internal/httpserver/deps"
	"github.com/MrSnakeDoc/jump/internal/httpserver/mw"
	"github.com/MrSnakeDoc/jump/internal/httpserver/routes"
	"github.com/MrSnakeDoc/jump/internal/logger"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	http    *http.Server
	logger  logger.Logger
	started time.Time
}

// New builds the HTTP server (router, middlewares, route registration).
func New(cfg *config.Config, loggerClient logger.Logger, d deps.Deps) *Server {
	r := chi.NewRouter()

	// --- Global middlewares (safe defaults)
	r.Use(middleware.GetHead)
	r.Use(middleware.RequestID)                // X-Request-ID on each request
	r.Use(middleware.Recoverer)                // never crash the process on panic
	r.Use(middleware.Timeout(2 * time.Second)) // per-request timeout (adjust if needed)
	r.Use(mw.Log(loggerClient))                // structured access logs (you'll implement)
	r.Use(mw.CORS())                           // optional: add if you expose publicly

	// Auto-register all routes under /api
	routes.RegisterAll(r, d)

	s := &http.Server{
		Addr:              cfg.ListenPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return &Server{
		http:    s,
		logger:  loggerClient,
		started: d.StartTime,
	}
}

// Start runs the HTTP server (blocks until error or shutdown).
func (s *Server) Start() error {
	s.logger.Infof("HTTP server listening on %s", s.http.Addr)
	err := s.http.ListenAndServe()
	// http.ErrServerClosed is expected on graceful shutdown.
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Stop gracefully shuts down the server with the provided context deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down...")
	return s.http.Shutdown(ctx)
}
