package geo

import "math"

func InitialBearing(a, b LatLon) float64 {
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}
	return bearing
}

func CourseChange(cog1, cog2 float64) float64 {
	diff := cog2 - cog1
	for diff > 180 {
		diff -= 360
	}
	for diff < -180 {
		diff += 360
	}
	if diff < 0 {
		return -diff
	}
	return diff
}

func MidPoint(a, b LatLon) LatLon {
	lat1 := a.Lat * math.Pi / 180
	lon1 := a.Lon * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	bx := math.Cos(lat2) * math.Cos(dLon)
	by := math.Cos(lat2) * math.Sin(dLon)

	lat3 := math.Atan2(
		math.Sin(lat1)+math.Sin(lat2),
		math.Sqrt((math.Cos(lat1)+bx)*(math.Cos(lat1)+bx)+by*by),
	)
	lon3 := lon1 + math.Atan2(by, math.Cos(lat1)+bx)

	return LatLon{
		Lat: lat3 * 180 / math.Pi,
		Lon: lon3 * 180 / math.Pi,
	}
}

func BoundingBoxContains(minLat, maxLat, minLon, maxLon float64, p LatLon) bool {
	return p.Lat >= minLat && p.Lat <= maxLat && p.Lon >= minLon && p.Lon <= maxLon
}

func Destination(p LatLon, bearingDeg, distKm float64) LatLon {
	if distKm == 0 {
		return p
	}
	if distKm < 0 {
		return Destination(p, bearingDeg+180, -distKm)
	}
	δ := distKm / earthRadiusKm
	θ := bearingDeg * math.Pi / 180
	φ1 := p.Lat * math.Pi / 180
	λ1 := p.Lon * math.Pi / 180
	sinφ1 := math.Sin(φ1)
	cosφ1 := math.Cos(φ1)
	sinδ := math.Sin(δ)
	cosδ := math.Cos(δ)
	sinθ := math.Sin(θ)
	cosθ := math.Cos(θ)
	sinφ2 := sinφ1*cosδ + cosφ1*sinδ*cosθ
	if sinφ2 > 1 {
		sinφ2 = 1
	}
	if sinφ2 < -1 {
		sinφ2 = -1
	}
	φ2 := math.Asin(sinφ2)
	y := sinθ * sinδ * cosφ1
	x := cosδ - sinφ1*sinφ2
	λ2 := λ1 + math.Atan2(y, x)
	return LatLon{
		Lat: φ2 * 180 / math.Pi,
		Lon: NormalizeLon(λ2 * 180 / math.Pi),
	}
}

func NormalizeLon(lon float64) float64 {
	for lon <= -180 {
		lon += 360
	}
	for lon > 180 {
		lon -= 360
	}
	return lon
}

func KnotsToKmPerHour(sog float64) float64 {
	return sog * 1.852
}

func KmToNauticalMiles(km float64) float64 {
	return km / 1.852
}

func PolygonArea(poly []LatLon) float64 {
	if len(poly) < 3 {
		return 0
	}
	var cLat, cLon float64
	for _, p := range poly {
		cLat += p.Lat
		cLon += p.Lon
	}
	cLat /= float64(len(poly))
	cLon /= float64(len(poly))

	cosLat := math.Cos(cLat * math.Pi / 180)
	kmPerDegLat := 111.32
	kmPerDegLon := 111.32 * cosLat

	n := len(poly)
	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		xi := (poly[i].Lon - cLon) * kmPerDegLon
		yi := (poly[i].Lat - cLat) * kmPerDegLat
		xj := (poly[j].Lon - cLon) * kmPerDegLon
		yj := (poly[j].Lat - cLat) * kmPerDegLat
		area += xi*yj - xj*yi
	}
	if area < 0 {
		area = -area
	}
	return area / 2
}
