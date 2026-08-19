package geo

import "math"

// InitialBearing returns the initial bearing (forward azimuth) in degrees
// from point a to point b along a great circle. The result is in [0, 360).
func InitialBearing(a, b LatLon) float64 {
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	// Normalize to [0, 360)
	if bearing < 0 {
		bearing += 360
	}
	return bearing
}

// CourseChange returns the absolute change in course between two
// consecutive headings in degrees. The result is in [0, 180].
func CourseChange(cog1, cog2 float64) float64 {
	diff := cog2 - cog1
	// Normalize to [-180, 180]
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

// MidPoint returns the geographic midpoint between two coordinates along
// a great circle.
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

// BoundingBoxContains checks if a point is within the given bounding box.
func BoundingBoxContains(minLat, maxLat, minLon, maxLon float64, p LatLon) bool {
	return p.Lat >= minLat && p.Lat <= maxLat && p.Lon >= minLon && p.Lon <= maxLon
}

// PolygonArea computes the approximate area of a polygon in square
// kilometers using the Shoelface formula on projected coordinates. This
// is a rough approximation suitable for small regions.
func PolygonArea(poly []LatLon) float64 {
	if len(poly) < 3 {
		return 0
	}
	// Convert to km using an equirectangular approximation centered on
	// the polygon's centroid.
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

	// Shoelace formula on projected coordinates
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
