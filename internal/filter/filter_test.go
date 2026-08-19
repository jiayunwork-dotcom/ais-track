package filter

import (
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestApplyBBox(t *testing.T) {
	recs := []parse.Record{
		{MMSI: "A", Lat: 10, Lon: 20, SOG: 5},
		{MMSI: "B", Lat: 50, Lon: 60, SOG: 8},
	}
	c := &Criteria{MinLat: 5, MaxLat: 15, MinLon: 15, MaxLon: 25}
	out := Apply(recs, c)
	if len(out) != 1 {
		t.Fatalf("got %d records, want 1", len(out))
	}
	if out[0].MMSI != "A" {
		t.Errorf("MMSI = %q, want A", out[0].MMSI)
	}
}

func TestDedupRemovesDuplicates(t *testing.T) {
	t0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	track := []parse.Record{
		{MMSI: "A", Timestamp: t0, Lat: 1, Lon: 2, SOG: 5},
		{MMSI: "A", Timestamp: t0.Add(10 * time.Second), Lat: 1, Lon: 2, SOG: 5},
		{MMSI: "A", Timestamp: t0.Add(60 * time.Second), Lat: 1.1, Lon: 2.1, SOG: 6},
	}
	cfg := DefaultDedupConfig()
	out := Dedup(track, cfg)
	if len(out) != 2 {
		t.Fatalf("got %d records, want 2 (middle should be deduped)", len(out))
	}
}

func TestValidateCoordinates(t *testing.T) {
	recs := []parse.Record{
		{MMSI: "A", Lat: 45, Lon: 90},
		{MMSI: "B", Lat: 100, Lon: 50}, // invalid lat
		{MMSI: "C", Lat: 0, Lon: 200},  // invalid lon
	}
	out := ValidateCoordinates(recs)
	if len(out) != 1 {
		t.Fatalf("got %d valid records, want 1", len(out))
	}
}
