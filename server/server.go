package server

import (
	"net/http"
	"safelyyou-assessment/devices"
)

// Server holds shared dependencies available to all handlers
type Server struct {
	store *devices.Store
}

// New creates a Server with the given device store
func New(store *devices.Store) *Server {
	return &Server{
		store: store,
	}
}

// Run registers routes and starts the HTTP server on the given address
func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	return http.ListenAndServe(addr, mux)
}

// registerRoutes wires URL patterns to their handler functions
func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", s.handleStats)
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", s.handleGetStats)
}
