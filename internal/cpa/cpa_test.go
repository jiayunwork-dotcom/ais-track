package cpa

import (
	"testing"
	"time"

	"ais-track/internal/parse"
)

func TestCPAHeadOn(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	own := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.00, Lon: 129.0, SOG: 12, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(20 * time.Minute), Lat: 35.04, Lon: 129.0, SOG: 12, COG: 0},
	}
	target := []parse.Record{
		{MMSI: "412000002", Timestamp: t0, Lat: 35.04, Lon: 129.0, SOG: 12, COG: 180},
		{MMSI: "412000002", Timestamp: t0.Add(20 * time.Minute), Lat: 35.00, Lon: 129.0, SOG: 12, COG: 180},
	}
	res, err := Compute(own, target, 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if res.DCPAKm > 0.15 {
		t.Fatalf("head-on DCPA %v, want near 0", res.DCPAKm)
	}
	mid := t0.Add(10 * time.Minute)
	dt := res.At.Sub(mid)
	if dt < 0 {
		dt = -dt
	}
	if dt > 2*time.Minute {
		t.Fatalf("CPA time %v, want near %v", res.At, mid)
	}
}

func TestCPAParallelOffset(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	own := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.00, Lon: 129.00, SOG: 12, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(20 * time.Minute), Lat: 35.04, Lon: 129.00, SOG: 12, COG: 0},
	}
	target := []parse.Record{
		{MMSI: "412000002", Timestamp: t0, Lat: 35.00, Lon: 129.01, SOG: 12, COG: 0},
		{MMSI: "412000002", Timestamp: t0.Add(20 * time.Minute), Lat: 35.04, Lon: 129.01, SOG: 12, COG: 0},
	}
	res, err := Compute(own, target, 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if res.DCPAKm < 0.7 || res.DCPAKm > 1.3 {
		t.Fatalf("parallel DCPA %v, want ~1.0 km", res.DCPAKm)
	}
}

func TestCPANoOverlap(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	own := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.00, Lon: 129.0, SOG: 12, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(10 * time.Minute), Lat: 35.02, Lon: 129.0, SOG: 12, COG: 0},
	}
	target := []parse.Record{
		{MMSI: "412000002", Timestamp: t0.Add(2 * time.Hour), Lat: 35.00, Lon: 129.0, SOG: 12, COG: 180},
		{MMSI: "412000002", Timestamp: t0.Add(2*time.Hour + 10*time.Minute), Lat: 34.98, Lon: 129.0, SOG: 12, COG: 180},
	}
	if _, err := Compute(own, target, 2*time.Hour); err == nil {
		t.Fatal("expected no-overlap error")
	}
}

func TestCPADoesNotInventAcrossGap(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	own := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.00, Lon: 129.0, SOG: 12, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(3 * time.Hour), Lat: 35.40, Lon: 129.0, SOG: 12, COG: 0},
	}
	target := []parse.Record{
		{MMSI: "412000002", Timestamp: t0.Add(30 * time.Minute), Lat: 35.10, Lon: 129.0, SOG: 12, COG: 180},
		{MMSI: "412000002", Timestamp: t0.Add(50 * time.Minute), Lat: 35.05, Lon: 129.0, SOG: 12, COG: 180},
	}
	if _, err := Compute(own, target, 2*time.Hour); err == nil {
		t.Fatal("CPA must not invent a closest approach across a reporting gap")
	}
}

func TestCPASameMMSI(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	own := []parse.Record{
		{MMSI: "412000001", Timestamp: t0, Lat: 35.00, Lon: 129.0, SOG: 12, COG: 0},
		{MMSI: "412000001", Timestamp: t0.Add(10 * time.Minute), Lat: 35.02, Lon: 129.0, SOG: 12, COG: 0},
	}
	if _, err := Compute(own, own, 2*time.Hour); err == nil {
		t.Fatal("same MMSI must be rejected")
	}
}

func TestRiskAlert(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	r := Result{DCPAKm: 0.1, At: t0.Add(5 * time.Minute)}
	got, err := Classify(r, t0, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got != RiskAlert {
		t.Fatalf("risk %q want alert", got)
	}
}

func TestRiskNoneWhenFar(t *testing.T) {
	t0 := time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)
	r := Result{DCPAKm: 4.0, At: t0.Add(5 * time.Minute)}
	got, err := Classify(r, t0, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got != RiskNone {
		t.Fatalf("risk %q want none", got)
	}
}
