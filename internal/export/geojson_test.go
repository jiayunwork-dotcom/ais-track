package export

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestTrackToGeoJSON(t *testing.T) {
	track := []parse.Record{
		{MMSI: "A", Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Lat: 1.0, Lon: 2.0, SOG: 5},
		{MMSI: "A", Timestamp: time.Date(2023, 1, 1, 0, 10, 0, 0, time.UTC), Lat: 1.1, Lon: 2.1, SOG: 6},
	}
	f := TrackToGeoJSON("A", track)
	if f.Type != "Feature" {
		t.Errorf("type = %q, want Feature", f.Type)
	}
	if f.Geometry.Type != "LineString" {
		t.Errorf("geometry type = %q, want LineString", f.Geometry.Type)
	}
	if f.Properties["mmsi"] != "A" {
		t.Errorf("mmsi = %v, want A", f.Properties["mmsi"])
	}
}

func TestWriteGeoJSON(t *testing.T) {
	fc := &FeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{Type: "Feature", Geometry: Geometry{Type: "Point", Coordinates: [2]float64{1, 2}}, Properties: map[string]interface{}{}},
		},
	}
	var buf bytes.Buffer
	if err := WriteGeoJSON(&buf, fc); err != nil {
		t.Fatalf("WriteGeoJSON: %v", err)
	}
	var decoded FeatureCollection
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.Type != "FeatureCollection" {
		t.Errorf("type = %q", decoded.Type)
	}
}
