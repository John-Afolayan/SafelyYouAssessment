package server

import (
	"net/http"
	"time"
	"encoding/json"
)

// heartbeatRequest represents the expected JSON body for POST /heartbeat
type heartbeatRequest struct {
    SentAt time.Time `json:"sent_at"`
}
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	deviceId := r.PathValue(("device_id")) // extract device_id val from path

	// return 404 if device_id is not found
	device, ok := s.store.Devices[deviceId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// parse json body
	var req heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// lock mutex and append timestamp
	device.AddHeartbeat(req.SentAt)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}