package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health code %d", rec.Code)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}" && !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("health body %q", rec.Body.String())
	}
}

func TestAnalyzeEndpoint(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{"max_sog":30,"records":[{"mmsi":"412000001","ts":"2023-06-01T08:00:00Z","lat":35.0,"lon":129.0,"sog":8.5,"cog":180},{"mmsi":"412000001","ts":"2023-06-01T08:10:00Z","lat":35.1,"lon":129.1,"sog":40.0,"cog":182}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("analyze code %d body %s", rec.Code, rec.Body.String())
	}
	var resp analyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Records != 2 {
		t.Fatalf("records %d", resp.Records)
	}
	if resp.SpeedingCount != 1 {
		t.Fatalf("speeding %d want 1", resp.SpeedingCount)
	}
}

func TestAnalyzeInvalidLat(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{"records":[{"mmsi":"412000001","ts":"2023-06-01T08:00:00Z","lat":95.0,"lon":129.0,"sog":8.5,"cog":180}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAnalyzeEmptyBody(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(nil))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCPAEndpoint(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{"own":[{"mmsi":"412000001","ts":"2023-06-01T08:00:00Z","lat":35.00,"lon":129.0,"sog":12,"cog":0},{"mmsi":"412000001","ts":"2023-06-01T08:20:00Z","lat":35.04,"lon":129.0,"sog":12,"cog":0}],"target":[{"mmsi":"412000002","ts":"2023-06-01T08:00:00Z","lat":35.04,"lon":129.0,"sog":12,"cog":180},{"mmsi":"412000002","ts":"2023-06-01T08:20:00Z","lat":35.00,"lon":129.0,"sog":12,"cog":180}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/cpa", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("cpa code %d body %s", rec.Code, rec.Body.String())
	}
	var resp cpaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.DCPAKm > 0.2 {
		t.Fatalf("head-on DCPA %v should be near zero", resp.DCPAKm)
	}
}

func TestCPANoOverlap(t *testing.T) {
	srv := New(DefaultConfig())
	raw := []byte(`{"own":[{"mmsi":"412000001","ts":"2023-06-01T08:00:00Z","lat":35.0,"lon":129.0,"sog":12,"cog":0},{"mmsi":"412000001","ts":"2023-06-01T08:10:00Z","lat":35.02,"lon":129.0,"sog":12,"cog":0}],"target":[{"mmsi":"412000002","ts":"2023-06-01T10:00:00Z","lat":35.0,"lon":129.0,"sog":12,"cog":180},{"mmsi":"412000002","ts":"2023-06-01T10:10:00Z","lat":34.98,"lon":129.0,"sog":12,"cog":180}]}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/cpa", bytes.NewReader(raw))
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAnalyzeMethodNotAllowed(t *testing.T) {
	srv := New(DefaultConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analyze", nil)
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
