package detect

import (
	"testing"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

func mkRec(mmsi string, sog, lat, lon float64) parse.Record {
	return parse.Record{
		MMSI:      mmsi,
		Timestamp: time.Now(),
		Lat:       lat,
		Lon:       lon,
		SOG:       sog,
	}
}

func TestSpeedingCount(t *testing.T) {
	track := []parse.Record{
		mkRec("A", 5, 0, 0),
		mkRec("A", 40, 0, 0),
		mkRec("A", 50, 0, 0),
		mkRec("A", 10, 0, 0),
	}
	if got := SpeedingCount(track, 30); got != 2 {
		t.Fatalf("SpeedingCount = %d, want 2", got)
	}
}

func TestAnomaliesNil(t *testing.T) {
	if got := Anomalies(nil, 30, nil); got != nil {
		t.Fatalf("Anomalies(nil) should return nil slice, got %d", len(got))
	}
}

func TestAnomaliesSpeeding(t *testing.T) {
	track := []parse.Record{
		mkRec("A", 5, 0, 0),
		mkRec("A", 40, 0, 0),
	}
	anoms := Anomalies(track, 30, nil)
	if len(anoms) != 1 {
		t.Fatalf("want 1 anomaly, got %d", len(anoms))
	}
	if anoms[0].Kind != "speeding" {
		t.Errorf("kind = %q, want speeding", anoms[0].Kind)
	}
}

func TestAnomaliesLoitering(t *testing.T) {
	port := []geo.LatLon{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 10},
		{Lat: 10, Lon: 10},
		{Lat: 10, Lon: 0},
	}
	track := []parse.Record{
		mkRec("A", 1, 5, 5),
		mkRec("A", 1, 5, 5),
		mkRec("A", 1, 5, 5), // 3rd consecutive in-port -> loitering
		mkRec("A", 1, 5, 5),
	}
	anoms := Anomalies(track, 30, port)
	count := 0
	for _, a := range anoms {
		if a.Kind == "loitering" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("want 1 loitering anomaly, got %d", count)
	}
}

func TestAnomaliesNoLoiterBelowThree(t *testing.T) {
	port := []geo.LatLon{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 10},
		{Lat: 10, Lon: 10},
		{Lat: 10, Lon: 0},
	}
	track := []parse.Record{
		mkRec("A", 1, 5, 5),
		mkRec("A", 1, 5, 5), // only 2 consecutive -> no loitering
	}
	anoms := Anomalies(track, 30, port)
	for _, a := range anoms {
		if a.Kind == "loitering" {
			t.Fatal("did not expect loitering for <3 consecutive in-port records")
		}
	}
}
