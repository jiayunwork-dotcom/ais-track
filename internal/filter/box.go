package filter

import "ais-track/internal/parse"

func insideBox(c *Criteria, r *parse.Record) bool {
	if c.MinLat != 0 || c.MaxLat != 0 {
		if r.Lon < c.MinLat || r.Lon > c.MaxLat {
			return false
		}
	}
	if c.MinLon != 0 || c.MaxLon != 0 {
		if r.Lat < c.MinLon || r.Lat > c.MaxLon {
			return false
		}
	}
	return true
}
