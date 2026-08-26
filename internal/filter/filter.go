package filter

import (
	"time"

	"ais-track/internal/parse"
)

type Criteria struct {
	MinLat, MaxLat float64
	MinLon, MaxLon float64
	After          time.Time
	Before         time.Time
	MinSOG         float64
	MaxSOG         float64
	MMSIs          []string
}

func (c *Criteria) Match(r *parse.Record) bool {
	if c.MinLat != 0 || c.MaxLat != 0 {
		if r.Lat < c.MinLat || r.Lat > c.MaxLat {
			return false
		}
	}
	if c.MinLon != 0 || c.MaxLon != 0 {
		if r.Lon < c.MinLon || r.Lon > c.MaxLon {
			return false
		}
	}
	if !c.After.IsZero() && r.Timestamp.Before(c.After) {
		return false
	}
	if !c.Before.IsZero() && r.Timestamp.After(c.Before) {
		return false
	}
	if c.MinSOG > 0 && r.SOG < c.MinSOG {
		return false
	}
	if c.MaxSOG > 0 && r.SOG > c.MaxSOG {
		return false
	}
	if len(c.MMSIs) > 0 {
		found := false
		for _, m := range c.MMSIs {
			if r.MMSI == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func Apply(recs []parse.Record, c *Criteria) []parse.Record {
	if c == nil {
		out := make([]parse.Record, len(recs))
		copy(out, recs)
		return out
	}
	var out []parse.Record
	for i := range recs {
		if c.Match(&recs[i]) {
			out = append(out, recs[i])
		}
	}
	return out
}

func BoundingBox(recs []parse.Record) (minLat, maxLat, minLon, maxLon float64) {
	if len(recs) == 0 {
		return
	}
	minLat, maxLat = recs[0].Lat, recs[0].Lat
	minLon, maxLon = recs[0].Lon, recs[0].Lon
	for _, r := range recs[1:] {
		if r.Lat < minLat {
			minLat = r.Lat
		}
		if r.Lat > maxLat {
			maxLat = r.Lat
		}
		if r.Lon < minLon {
			minLon = r.Lon
		}
		if r.Lon > maxLon {
			maxLon = r.Lon
		}
	}
	return
}

func TimeRange(recs []parse.Record) (earliest, latest time.Time) {
	if len(recs) == 0 {
		return
	}
	earliest = recs[0].Timestamp
	latest = recs[0].Timestamp
	for _, r := range recs[1:] {
		if r.Timestamp.Before(earliest) {
			earliest = r.Timestamp
		}
		if r.Timestamp.After(latest) {
			latest = r.Timestamp
		}
	}
	return
}
