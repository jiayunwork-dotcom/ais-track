package interp

import (
	"math"
	"testing"
	"time"

	"ais-track/internal/parse"
)

func sampleLeg() []parse.Record {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	return []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.0, Lon: 129.0, SOG: 10, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(10 * time.Minute), Lat: 35.1, Lon: 129.0, SOG: 10, COG: 0},
	}
}

func TestInterpAtMidLeg(t *testing.T) {
	track := sampleLeg()
	mid := track[0].Timestamp.Add(5 * time.Minute)
	got, err := At(track, mid, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Lat-35.05) > 1e-9 {
		t.Fatalf("mid lat %v want 35.05", got.Lat)
	}
	if got.Lon != 129.0 {
		t.Fatalf("lon %v", got.Lon)
	}
}

func TestInterpRejectsGap(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	track := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.0, Lon: 129.0, SOG: 10, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(3 * time.Hour), Lat: 35.5, Lon: 129.0, SOG: 10, COG: 0},
	}
	_, err := At(track, t0.Add(90*time.Minute), 2*time.Hour)
	if err == nil {
		t.Fatal("expected gap rejection")
	}
}

func TestInterpOutsideRange(t *testing.T) {
	track := sampleLeg()
	_, err := At(track, track[0].Timestamp.Add(-time.Minute), time.Hour)
	if err == nil {
		t.Fatal("expected outside-range error")
	}
}

func TestDeadReckonNorth(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	r := parse.Record{MMSI: "412000001", Timestamp: t0, Lat: 0, Lon: 0, SOG: 60, COG: 0}
	got, err := DeadReckon(r, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Lat-1) > 0.02 {
		t.Fatalf("dead-reckon lat %v want ~1 deg", got.Lat)
	}
	if math.Abs(got.Lon) > 0.02 {
		t.Fatalf("dead-reckon lon %v want ~0", got.Lon)
	}
}

func TestSampleEveryHitsEndpoints(t *testing.T) {
	track := sampleLeg()
	got, err := SampleEvery(track, 5*time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 3 {
		t.Fatalf("samples %d", len(got))
	}
	if !got[0].Timestamp.Equal(track[0].Timestamp) {
		t.Fatal("missing start sample")
	}
}
