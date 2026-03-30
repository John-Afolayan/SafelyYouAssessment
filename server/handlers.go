package server

import "net/http"

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}