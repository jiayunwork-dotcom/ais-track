package voyage

import (
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestSegmentGap(t *testing.T) {
	t0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	track := []parse.Record{
		{MMSI: "A", Timestamp: t0, Lat: 0, Lon: 0, SOG: 5},
		{MMSI: "A", Timestamp: t0.Add(30 * time.Minute), Lat: 0.1, Lon: 0, SOG: 5},
		// 3-hour gap -> new voyage
		{MMSI: "A", Timestamp: t0.Add(4 * time.Hour), Lat: 1, Lon: 0, SOG: 8},
		{MMSI: "A", Timestamp: t0.Add(5 * time.Hour), Lat: 1.1, Lon: 0, SOG: 8},
	}
	cfg := DefaultConfig()
	voyages := Segment(track, cfg)
	if len(voyages) != 2 {
		t.Fatalf("got %d voyages, want 2", len(voyages))
	}
	if voyages[0].RecordCount() != 2 {
		t.Errorf("voyage 0: %d records, want 2", voyages[0].RecordCount())
	}
	if voyages[1].RecordCount() != 2 {
		t.Errorf("voyage 1: %d records, want 2", voyages[1].RecordCount())
	}
}

func TestSegmentEmpty(t *testing.T) {
	if got := Segment(nil, DefaultConfig()); got != nil {
		t.Fatalf("Segment(nil) = %v, want nil", got)
	}
}

func TestIsAnchored(t *testing.T) {
	t0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	track := []parse.Record{
		{MMSI: "A", Timestamp: t0, Lat: 35.0, Lon: 129.0},
		{MMSI: "A", Timestamp: t0.Add(10 * time.Minute), Lat: 35.0001, Lon: 129.0001},
	}
	v := Voyage{Records: track}
	if !IsAnchored(&v, 0.5) {
		t.Error("expected anchored for nearby points")
	}
}
