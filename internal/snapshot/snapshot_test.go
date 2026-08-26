package snapshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ais-track/internal/detect"
	"ais-track/internal/parse"
	"ais-track/internal/voyage"
)

func sampleTrack() []parse.Record {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	return []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.0, Lon: 129.0, SOG: 8.5, COG: 180},
		{MMSI: "412000001", Timestamp: t0.Add(10 * time.Minute), Lat: 35.1, Lon: 129.1, SOG: 40, COG: 182},
		{MMSI: "412000001", Timestamp: t0.Add(4 * time.Hour), Lat: 36.0, Lon: 129.5, SOG: 9, COG: 90},
		{MMSI: "412000001", Timestamp: t0.Add(4*time.Hour + 10*time.Minute), Lat: 36.05, Lon: 129.55, SOG: 9, COG: 92},
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	recs := sampleTrack()
	dir := t.TempDir()
	path := filepath.Join(dir, "track.json")
	if err := WriteFile(path, recs); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Matches(recs, got, 1e-9, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyFileRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(path); err == nil {
		t.Fatal("expected empty-file error")
	}
}

func TestTruncatedRejected(t *testing.T) {
	recs := sampleTrack()
	dir := t.TempDir()
	path := filepath.Join(dir, "trunc.json")
	if err := WriteFile(path, recs); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 40 {
		t.Fatal("snapshot too small to truncate")
	}
	if err := os.WriteFile(path, b[:len(b)/2], 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(path); err == nil {
		t.Fatal("expected truncated-file error")
	}
}

func TestRefuseEmptyWrite(t *testing.T) {
	dir := t.TempDir()
	if err := WriteFile(filepath.Join(dir, "none.json"), nil); err == nil {
		t.Fatal("expected empty write error")
	}
}

func TestSnapshotPreservesAnomalies(t *testing.T) {
	recs := sampleTrack()
	dir := t.TempDir()
	path := filepath.Join(dir, "anom.json")
	if err := WriteFile(path, recs); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := detect.SpeedingCount(recs, 30)
	after := detect.SpeedingCount(got, 30)
	if before != after || before != 1 {
		t.Fatalf("speeding before=%d after=%d", before, after)
	}
}

func TestSnapshotPreservesVoyageSegments(t *testing.T) {
	recs := sampleTrack()
	dir := t.TempDir()
	path := filepath.Join(dir, "voy.json")
	if err := WriteFile(path, recs); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := voyage.Segment(recs, voyage.DefaultConfig())
	after := voyage.Segment(got, voyage.DefaultConfig())
	if len(before) != 2 || len(after) != 2 {
		t.Fatalf("voyage count before=%d after=%d", len(before), len(after))
	}
	if before[0].RecordCount() != after[0].RecordCount() {
		t.Fatalf("first voyage records %d vs %d", before[0].RecordCount(), after[0].RecordCount())
	}
}
