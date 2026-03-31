package server

import (
	"encoding/json"
	"net/http"
	"safelyyou-assessment/devices"
	"time"
)

// ---------- Request / Response types ----------

type heartbeatRequest struct {
	SentAt time.Time `json:"sent_at"`
}

type uploadStatsRequest struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"`
}

type deviceStatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

type errorResponse struct {
	Msg string `json:"msg"`
}

// ---------- Helpers ----------

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// lookupDevice extracts device_id from the path and looks it up.
// Returns nil and writes a 404 JSON response if the device is not found.
func (s *Server) lookupDevice(w http.ResponseWriter, r *http.Request) *devices.Device {
	id := r.PathValue("device_id")
	device, ok := s.store.Lookup(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Msg: "device not found: " + id,
		})
		return nil
	}
	return device
}

// ---------- Handlers ----------

// handleHeartbeat registers a heartbeat for a device.
//
//	POST /api/v1/devices/{device_id}/heartbeat
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	device := s.lookupDevice(w, r)
	if device == nil {
		return
	}

	var req heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Msg: "invalid JSON: " + err.Error()})
		return
	}

	if req.SentAt.IsZero() {
		writeJSON(w, http.StatusBadRequest, errorResponse{Msg: "sent_at is required"})
		return
	}

	device.AddHeartbeat(req.SentAt)
	w.WriteHeader(http.StatusNoContent)
}

// handlePostStats records an upload stat for a device.
//
//	POST /api/v1/devices/{device_id}/stats
func (s *Server) handlePostStats(w http.ResponseWriter, r *http.Request) {
	device := s.lookupDevice(w, r)
	if device == nil {
		return
	}

	var req uploadStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Msg: "invalid JSON: " + err.Error()})
		return
	}

	if req.UploadTime < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Msg: "upload_time must be non-negative"})
		return
	}

	device.AddUpload(time.Duration(req.UploadTime))
	w.WriteHeader(http.StatusNoContent)
}

// handleGetStats returns computed device statistics.
//
//	GET /api/v1/devices/{device_id}/stats
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	device := s.lookupDevice(w, r)
	if device == nil {
		return
	}

	writeJSON(w, http.StatusOK, deviceStatsResponse{
		Uptime:        devices.CalculateUptime(device),
		AvgUploadTime: devices.CalculateAvgUploadTime(device),
	})
}
