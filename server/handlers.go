package server

import (
	"encoding/json"
	"net/http"
	"safelyyou-assessment/devices"
	"time"
)

// heartbeatRequest represents the expected JSON body for POST /heartbeat
type heartbeatRequest struct {
	SentAt time.Time `json:"sent_at"`
}

// uploadStatsRequest represents the expected JSON body for POST /stats
type uploadStatsRequest struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"`
}

// deviceStatsResponse represents the JSON response body for GET /stats
type deviceStatsResponse struct {
    Uptime        float64 `json:"uptime"`
    AvgUploadTime string  `json:"avg_upload_time"`
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
	deviceId := r.PathValue(("device_id")) // extract device_id val from path

	// return 404 if device_id is not found
	device, ok := s.store.Devices[deviceId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// parse json body
	var req uploadStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// lock mutex and append upload time
	device.AddStat(time.Duration(req.UploadTime))

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	deviceId := r.PathValue(("device_id")) // extract device_id val from path

	// return 404 if device_id is not found
	device, ok := s.store.Devices[deviceId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// calculate stats
	uptime := devices.CalculateUptime(device)
	avgUploadTime := devices.CalculateAvgUploadTime(device)

	// write JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deviceStatsResponse {
		Uptime: uptime,
		AvgUploadTime: avgUploadTime,
	})

}
