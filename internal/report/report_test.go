package report

import (
	"bytes"
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestGenerateBasic(t *testing.T) {
	t0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	groups := map[string][]parse.Record{
		"V1": {
			{MMSI: "V1", Timestamp: t0, Lat: 0, Lon: 0, SOG: 5, COG: 90},
			{MMSI: "V1", Timestamp: t0.Add(10 * time.Minute), Lat: 0.1, Lon: 0, SOG: 40, COG: 95},
		},
	}
	fr := Generate(groups, 30, nil)
	if fr.Fleet.VesselCount != 1 {
		t.Errorf("VesselCount = %d, want 1", fr.Fleet.VesselCount)
	}
	if len(fr.Vessels) != 1 {
		t.Fatalf("Vessels = %d, want 1", len(fr.Vessels))
	}
	// Should have at least 1 anomaly (speeding at SOG=40 > 30)
	if len(fr.Vessels[0].Anomalies) == 0 {
		t.Error("expected at least one anomaly for speeding vessel")
	}
}

func TestWriteText(t *testing.T) {
	t0 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	groups := map[string][]parse.Record{
		"V1": {
			{MMSI: "V1", Timestamp: t0, Lat: 0, Lon: 0, SOG: 5, COG: 90},
		},
	}
	fr := Generate(groups, 30, nil)
	var buf bytes.Buffer
	err := WriteText(&buf, fr)
	if err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}
