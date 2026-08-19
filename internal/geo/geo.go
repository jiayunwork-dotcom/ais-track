// Package geo provides geographic helpers for AIS analysis: great-circle
// distance and point-in-polygon tests.
package geo

import (
	"math"

	"ais-track/internal/parse"
)

// LatLon is a geographic coordinate in decimal degrees.
type LatLon struct {
	Lat, Lon float64
}

const earthRadiusKm = 6371.0

// Haversine returns the great-circle distance in kilometers between a and b.
func Haversine(a, b LatLon) float64 {
	dLat := (b.Lat - a.Lat) * math.Pi / 180.0
	dLon := (b.Lon - a.Lon) * math.Pi / 180.0
	lat1 := a.Lat * math.Pi / 180.0
	lat2 := b.Lat * math.Pi / 180.0

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
	return earthRadiusKm * c
}

// PointInPolygon reports whether p lies inside poly using ray casting.
// A nil or empty polygon yields false.
func PointInPolygon(p LatLon, poly []LatLon) bool {
	n := len(poly)
	if n == 0 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := poly[i].Lon, poly[i].Lat
		xj, yj := poly[j].Lon, poly[j].Lat
		// Intersection test: p.Lat strictly between the edge's latitudes, and
		// the ray to the left of the edge. yi != yj is guaranteed by the
		// boolean condition, so the division is safe.
		intersect := ((yi > p.Lat) != (yj > p.Lat)) &&
			(p.Lon < (xj-xi)*(p.Lat-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

// LegDistance returns the great-circle distance in km between two AIS records.
func LegDistance(a, b parse.Record) float64 {
	return Haversine(
		LatLon{Lat: a.Lat, Lon: a.Lon},
		LatLon{Lat: b.Lat, Lon: b.Lon},
	)
}
