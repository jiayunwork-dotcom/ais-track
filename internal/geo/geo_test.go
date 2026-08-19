package geo

import (
	"math"
	"testing"

	"ais-track/internal/parse"
)

func TestHaversineZero(t *testing.T) {
	a := LatLon{Lat: 10.0, Lon: 20.0}
	if d := Haversine(a, a); d != 0 {
		t.Fatalf("distance to self = %v, want 0", d)
	}
}

func TestHaversineKnown(t *testing.T) {
	a := LatLon{Lat: 0, Lon: 0}
	b := LatLon{Lat: 1, Lon: 0}
	d := Haversine(a, b)
	if math.Abs(d-111.19) > 1.0 {
		t.Fatalf("distance = %v, want ~111.19", d)
	}
}

func TestPointInPolygonSquare(t *testing.T) {
	square := []LatLon{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 10},
		{Lat: 10, Lon: 10},
		{Lat: 10, Lon: 0},
	}
	if !PointInPolygon(LatLon{Lat: 5, Lon: 5}, square) {
		t.Error("point inside square should be true")
	}
	if PointInPolygon(LatLon{Lat: 15, Lon: 5}, square) {
		t.Error("point outside square should be false")
	}
}

func TestPointInPolygonNil(t *testing.T) {
	if PointInPolygon(LatLon{Lat: 5, Lon: 5}, nil) {
		t.Error("nil polygon should return false")
	}
	if PointInPolygon(LatLon{Lat: 5, Lon: 5}, []LatLon{}) {
		t.Error("empty polygon should return false")
	}
}

func TestLegDistance(t *testing.T) {
	a := parse.Record{Lat: 0, Lon: 0}
	b := parse.Record{Lat: 1, Lon: 0}
	if math.Abs(LegDistance(a, b)-111.19) > 1.0 {
		t.Fatalf("leg distance = %v, want ~111.19", LegDistance(a, b))
	}
}
