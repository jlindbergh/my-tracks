package api

import (
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	srv := New(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	srv.writeJSON(rec, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusOK, map[string]int{"a": 1})

	if rec.Code != http.StatusOK || rec.Body.String() != "{\"a\":1}\n" {
		t.Errorf("got %d %q", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestWriteJSONUnencodableValue(t *testing.T) {
	srv := New(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()
	srv.writeJSON(rec, httptest.NewRequest(http.MethodGet, "/", nil), http.StatusOK, map[string]float64{"distanceM": math.NaN()})

	// Not a 200 with an empty body, which the frontend cannot parse.
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "application/json" {
		t.Errorf("Content-Type = %q for an error that is not JSON", ct)
	}
}
