package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"safelyyou-assessment/devices"
	"syscall"
	"time"
)

// Server holds shared dependencies and the underlying HTTP server.
type Server struct {
	store  *devices.Store
	http   *http.Server
}

// New creates a Server with the given device store.
func New(store *devices.Store) *Server {
	return &Server{store: store}
}

// Run registers routes, starts the HTTP server, and blocks until a shutdown
// signal (SIGINT/SIGTERM) is received. Outstanding requests are given up to
// 5 seconds to complete before the server is forcibly stopped.
func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.http = &http.Server{
		Addr:         addr,
		Handler:      withRecovery(withBodyLimit(withRequestLogging(mux))),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// channel to capture server startup errors
	errCh := make(chan error, 1)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Printf("received signal %v, shutting down gracefully…", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}

// registerRoutes wires URL patterns to their handler functions.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", s.handlePostStats)
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", s.handleGetStats)

	// health check endpoint — useful for liveness probes in production
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}
