package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ais-track/internal/cpa"
	"ais-track/internal/detect"
	"ais-track/internal/parse"
)

type Server struct {
	mux  *http.ServeMux
	addr string
}

type Config struct {
	Addr string
}

type fixJSON struct {
	MMSI string  `json:"mmsi"`
	TS   string  `json:"ts"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	SOG  float64 `json:"sog"`
	COG  float64 `json:"cog"`
}

type analyzeRequest struct {
	MaxSOG  float64   `json:"max_sog"`
	Records []fixJSON `json:"records"`
}

type analyzeResponse struct {
	Records       int           `json:"records"`
	SpeedingCount int           `json:"speeding_count"`
	Anomalies     []anomalyJSON `json:"anomalies"`
}

type anomalyJSON struct {
	Kind   string `json:"kind"`
	At     string `json:"at"`
	Detail string `json:"detail"`
}

type cpaRequest struct {
	Own       []fixJSON `json:"own"`
	Target    []fixJSON `json:"target"`
	MaxGapSec float64   `json:"max_gap_sec"`
}

type cpaResponse struct {
	OwnMMSI    string  `json:"own_mmsi"`
	TargetMMSI string  `json:"target_mmsi"`
	DCPAKm     float64 `json:"dcpa_km"`
	At         string  `json:"at"`
	OwnLat     float64 `json:"own_lat"`
	OwnLon     float64 `json:"own_lon"`
	TargetLat  float64 `json:"target_lat"`
	TargetLon  float64 `json:"target_lon"`
}

func DefaultConfig() Config {
	return Config{Addr: ":8080"}
}

func New(cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	s := &Server{mux: http.NewServeMux(), addr: cfg.Addr}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) Addr() string { return s.addr }

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/analyze", s.handleAnalyze)
	s.mux.HandleFunc("/api/cpa", s.handleCPA)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req analyzeRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	recs, err := decodeFixes(req.Records)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if len(recs) == 0 {
		httpError(w, http.StatusUnprocessableEntity, "no records")
		return
	}
	maxSOG := req.MaxSOG
	if maxSOG <= 0 {
		maxSOG = 30
	}
	anoms := detect.Anomalies(recs, maxSOG, nil)
	out := make([]anomalyJSON, 0, len(anoms))
	for _, a := range anoms {
		out = append(out, anomalyJSON{
			Kind:   a.Kind,
			At:     a.At.UTC().Format(time.RFC3339),
			Detail: a.Detail,
		})
	}
	writeJSON(w, analyzeResponse{
		Records:       len(recs),
		SpeedingCount: detect.SpeedingCount(recs, maxSOG),
		Anomalies:     out,
	})
}

func (s *Server) handleCPA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req cpaRequest
	if err := readJSON(r, &req); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	own, err := decodeFixes(req.Own)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	target, err := decodeFixes(req.Target)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	gap := time.Duration(req.MaxGapSec * float64(time.Second))
	if gap <= 0 {
		gap = 2 * time.Hour
	}
	res, err := cpa.Compute(own, target, gap)
	if err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, cpaResponse{
		OwnMMSI:    res.OwnMMSI,
		TargetMMSI: res.TargetMMSI,
		DCPAKm:     res.DCPAKm,
		At:         res.At.UTC().Format(time.RFC3339),
		OwnLat:     res.OwnLat,
		OwnLon:     res.OwnLon,
		TargetLat:  res.TargetLat,
		TargetLon:  res.TargetLon,
	})
}

func decodeFixes(in []fixJSON) ([]parse.Record, error) {
	out := make([]parse.Record, 0, len(in))
	for i, f := range in {
		ts, err := time.Parse(time.RFC3339, f.TS)
		if err != nil {
			return nil, fmt.Errorf("record %d timestamp: %w", i, err)
		}
		rec := parse.Record{
			MMSI:      f.MMSI,
			Timestamp: ts,
			Lat:       f.Lat,
			Lon:       f.Lon,
			SOG:       f.SOG,
			COG:       f.COG,
		}
		out = append(out, rec)
	}
	kept := parse.FilterValid(out, time.Time{})
	return kept, nil
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	if len(body) == 0 {
		return fmt.Errorf("empty request body")
	}
	return json.Unmarshal(body, v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
