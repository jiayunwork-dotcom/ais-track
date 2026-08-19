package geo

import (
	"math"

	"ais-track/internal/parse"
)

// TrackDistance returns the total distance in km along a track (sum of
// consecutive leg distances).
func TrackDistance(track []parse.Record) float64 {
	if len(track) < 2 {
		return 0
	}
	total := 0.0
	for i := 1; i < len(track); i++ {
		total += LegDistance(track[i-1], track[i])
	}
	return total
}

// StraightLineDistance returns the great-circle distance between the first
// and last record of a track (displacement, not path distance).
func StraightLineDistance(track []parse.Record) float64 {
	if len(track) < 2 {
		return 0
	}
	a := LatLon{Lat: track[0].Lat, Lon: track[0].Lon}
	b := LatLon{Lat: track[len(track)-1].Lat, Lon: track[len(track)-1].Lon}
	return Haversine(a, b)
}

// Sinuosity returns the ratio of track distance to straight-line
// distance. A value of 1.0 means the vessel traveled in a perfect
// straight line; higher values indicate more winding paths.
func Sinuosity(track []parse.Record) float64 {
	straight := StraightLineDistance(track)
	if straight == 0 {
		return 0
	}
	return TrackDistance(track) / straight
}

// ClosestApproach returns the minimum distance in km between any record
// in track and the given point.
func ClosestApproach(track []parse.Record, target LatLon) float64 {
	if len(track) == 0 {
		return 0
	}
	min := math.MaxFloat64
	for _, r := range track {
		d := Haversine(LatLon{Lat: r.Lat, Lon: r.Lon}, target)
		if d < min {
			min = d
		}
	}
	return min
}

// MaxDeviation returns the maximum distance in km that any record deviates
// from the straight line between the first and last records. This is a
// simple measure of track curvature.
func MaxDeviation(track []parse.Record) float64 {
	if len(track) < 3 {
		return 0
	}
	// For each intermediate point, compute distance to the great-circle
	// line from start to end. We approximate by computing distance to
	// the midpoint's lat/lon projected line.
	start := LatLon{Lat: track[0].Lat, Lon: track[0].Lon}
	end := LatLon{Lat: track[len(track)-1].Lat, Lon: track[len(track)-1].Lon}
	mid := MidPoint(start, end)

	maxDev := 0.0
	for _, r := range track[1 : len(track)-1] {
		pt := LatLon{Lat: r.Lat, Lon: r.Lon}
		// Distance from point to midpoint minus half the straight-line
		// distance gives a rough deviation estimate.
		dToMid := Haversine(pt, mid)
		halfStraight := Haversine(start, end) / 2
		dev := math.Abs(dToMid - halfStraight)
		if dev > maxDev {
			maxDev = dev
		}
	}
	return maxDev
}
