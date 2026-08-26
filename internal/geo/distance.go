package geo

import (
	"math"

	"ais-track/internal/parse"
)

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

func StraightLineDistance(track []parse.Record) float64 {
	if len(track) < 2 {
		return 0
	}
	a := LatLon{Lat: track[0].Lat, Lon: track[0].Lon}
	b := LatLon{Lat: track[len(track)-1].Lat, Lon: track[len(track)-1].Lon}
	return Haversine(a, b)
}

func Sinuosity(track []parse.Record) float64 {
	straight := StraightLineDistance(track)
	if straight == 0 {
		return 0
	}
	return TrackDistance(track) / straight
}

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

func MaxDeviation(track []parse.Record) float64 {
	if len(track) < 3 {
		return 0
	}
	start := LatLon{Lat: track[0].Lat, Lon: track[0].Lon}
	end := LatLon{Lat: track[len(track)-1].Lat, Lon: track[len(track)-1].Lon}
	mid := MidPoint(start, end)

	maxDev := 0.0
	for _, r := range track[1 : len(track)-1] {
		pt := LatLon{Lat: r.Lat, Lon: r.Lon}
		dToMid := Haversine(pt, mid)
		halfStraight := Haversine(start, end) / 2
		dev := math.Abs(dToMid - halfStraight)
		if dev > maxDev {
			maxDev = dev
		}
	}
	return maxDev
}
