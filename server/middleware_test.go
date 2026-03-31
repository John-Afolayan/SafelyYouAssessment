package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithBodyLimit_RejectsOversized(t *testing.T) {
	// handler that reads the body
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 2<<20) // 2 MB buffer
		_, err := r.Body.Read(buf)
		if err == nil {
			t.Error("expected error reading oversized body")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := withBodyLimit(inner)

	// send a body larger than 1 MB
	bigBody := strings.NewReader(strings.Repeat("x", 2<<20))
	req := httptest.NewRequest("POST", "/", bigBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
}

func TestWithRecovery_CatchesPanic(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := withRecovery(inner)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// should not propagate the panic
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", w.Code)
	}
}
