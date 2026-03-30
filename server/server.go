package server

import (
	"net/http"
	"safelyyou-assessment/devices"
)

type Server struct {
	store *devices.Store
}

func New(store *devices.Store) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", s.handleStats)
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", s.handleGetStats)
}
