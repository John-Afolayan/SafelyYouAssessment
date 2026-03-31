package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"safelyyou-assessment/devices"
)

// newTestServer creates a Server with one registered device.
func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	store := devices.NewStore()
	store.Register("aa-bb-cc")
	return New(store), "aa-bb-cc"
}

func TestHeartbeat_Success(t *testing.T) {
	srv, id := newTestServer(t)

	body, _ := json.Marshal(heartbeatRequest{SentAt: time.Now()})
	req := httptest.NewRequest("POST", "/api/v1/devices/"+id+"/heartbeat", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", srv.handleHeartbeat)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHeartbeat_DeviceNotFound(t *testing.T) {
	srv, _ := newTestServer(t)

	body, _ := json.Marshal(heartbeatRequest{SentAt: time.Now()})
	req := httptest.NewRequest("POST", "/api/v1/devices/unknown/heartbeat", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", srv.handleHeartbeat)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp errorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Msg == "" {
		t.Error("expected error message in 404 response body")
	}
}

func TestHeartbeat_InvalidJSON(t *testing.T) {
	srv, id := newTestServer(t)

	req := httptest.NewRequest("POST", "/api/v1/devices/"+id+"/heartbeat", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/devices/{device_id}/heartbeat", srv.handleHeartbeat)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPostStats_Success(t *testing.T) {
	srv, id := newTestServer(t)

	body, _ := json.Marshal(uploadStatsRequest{SentAt: time.Now(), UploadTime: 5_000_000_000})
	req := httptest.NewRequest("POST", "/api/v1/devices/"+id+"/stats", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", srv.handlePostStats)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestPostStats_NegativeUploadTime(t *testing.T) {
	srv, id := newTestServer(t)

	body, _ := json.Marshal(uploadStatsRequest{SentAt: time.Now(), UploadTime: -1})
	req := httptest.NewRequest("POST", "/api/v1/devices/"+id+"/stats", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/devices/{device_id}/stats", srv.handlePostStats)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetStats_WithData(t *testing.T) {
	srv, id := newTestServer(t)

	device, _ := srv.store.Lookup(id)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= 10; i++ {
		device.AddHeartbeat(now.Add(time.Duration(i) * time.Minute))
	}
	device.AddUpload(2 * time.Second)
	device.AddUpload(4 * time.Second)

	req := httptest.NewRequest("GET", "/api/v1/devices/"+id+"/stats", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", srv.handleGetStats)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp deviceStatsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Uptime == 0 {
		t.Error("expected non-zero uptime")
	}
	if resp.AvgUploadTime != "3s" {
		t.Errorf("expected avg_upload_time '3s', got %q", resp.AvgUploadTime)
	}
}

func TestGetStats_NoData(t *testing.T) {
	srv, id := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/devices/"+id+"/stats", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/devices/{device_id}/stats", srv.handleGetStats)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp deviceStatsResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Uptime != 0 {
		t.Errorf("expected 0 uptime, got %f", resp.Uptime)
	}
	if resp.AvgUploadTime != "0s" {
		t.Errorf("expected '0s', got %q", resp.AvgUploadTime)
	}
}

func TestHealthCheck(t *testing.T) {
	store := devices.NewStore()
	srv := New(store)

	mux := http.NewServeMux()
	srv.registerRoutes(mux)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
