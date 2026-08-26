package stats

import (
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestComputeBasic(t *testing.T) {
	track := []parse.Record{
		{MMSI: "X", Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 0, Lon: 0, SOG: 5},
		{MMSI: "X", Timestamp: time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC), Lat: 1, Lon: 0, SOG: 10},
		{MMSI: "X", Timestamp: time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC), Lat: 2, Lon: 0, SOG: 15},
	}
	s := Compute(track)
	if s.RecordCount != 3 {
		t.Errorf("RecordCount = %d, want 3", s.RecordCount)
	}
	if s.MaxSOG != 15 {
		t.Errorf("MaxSOG = %f, want 15", s.MaxSOG)
	}
	if s.MinSOG != 5 {
		t.Errorf("MinSOG = %f, want 5", s.MinSOG)
	}
	if s.TotalDistKm < 200 {
		t.Errorf("TotalDistKm = %f, expected > 200", s.TotalDistKm)
	}
}

func TestHistogramPercentile(t *testing.T) {
	h := NewHistogram(0, 50, 10)
	for i := 0; i < 100; i++ {
		h.Add(float64(i % 50))
	}
	p50 := h.Percentile(0.5)
	if p50 < 20 || p50 > 30 {
		t.Errorf("p50 = %f, expected near 25", p50)
	}
}
