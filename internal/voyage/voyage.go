// Package voyage segments a vessel's AIS track into discrete voyages based
// on temporal gaps and spatial anchoring. A voyage begins when the vessel
// departs an anchor zone and ends when it arrives at the next anchor zone
// or when there is a reporting gap exceeding the configured threshold.
package voyage

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// Voyage represents one contiguous segment of a vessel's movement.
type Voyage struct {
	MMSI      string
	StartTime time.Time
	EndTime   time.Time
	Records   []parse.Record
	DistanceKm float64
	MaxSOG    float64
	AvgSOG    float64
}

// Duration returns the elapsed time of the voyage.
func (v *Voyage) Duration() time.Duration {
	return v.EndTime.Sub(v.StartTime)
}

// RecordCount returns the number of AIS fixes in this voyage.
func (v *Voyage) RecordCount() int {
	return len(v.Records)
}

// Config controls the voyage segmentation algorithm.
type Config struct {
	// GapThreshold is the maximum allowed time between consecutive records
	// before a new voyage is started.
	GapThreshold time.Duration
	// AnchorRadius is the radius in km around a point within which the
	// vessel is considered stationary (anchored).
	AnchorRadius float64
	// MinSOGMoving is the minimum speed over ground to consider the vessel
	// underway.
	MinSOGMoving float64
}

// DefaultConfig returns sensible defaults for voyage segmentation.
func DefaultConfig() Config {
	return Config{
		GapThreshold: 2 * time.Hour,
		AnchorRadius: 0.5,
		MinSOGMoving: 1.0,
	}
}

// Segment splits a sorted track into voyages. The track must be sorted by
// timestamp (caller responsibility). It returns nil for an empty track.
func Segment(track []parse.Record, cfg Config) []Voyage {
	if len(track) == 0 {
		return nil
	}

	var voyages []Voyage
	var current []parse.Record
	current = append(current, track[0])

	for i := 1; i < len(track); i++ {
		prev := track[i-1]
		cur := track[i]

		gap := cur.Timestamp.Sub(prev.Timestamp)
		if gap > cfg.GapThreshold {
			// Time gap: close the current voyage and start a new one
			voyages = append(voyages, buildVoyage(current))
			current = current[:0]
		}

		current = append(current, cur)
	}

	if len(current) > 0 {
		voyages = append(voyages, buildVoyage(current))
	}
	return voyages
}

// buildVoyage computes the derived statistics for a voyage segment.
func buildVoyage(recs []parse.Record) Voyage {
	if len(recs) == 0 {
		return Voyage{}
	}
	v := Voyage{
		MMSI:      recs[0].MMSI,
		StartTime: recs[0].Timestamp,
		EndTime:   recs[len(recs)-1].Timestamp,
		Records:   recs,
	}

	var totalSOG float64
	for i, r := range recs {
		totalSOG += r.SOG
		if r.SOG > v.MaxSOG {
			v.MaxSOG = r.SOG
		}
		if i > 0 {
			v.DistanceKm += geo.LegDistance(recs[i-1], r)
		}
	}
	if len(recs) > 0 {
		v.AvgSOG = totalSOG / float64(len(recs))
	}
	return v
}

// IsAnchored reports whether the vessel stayed within anchorRadius of its
// first position for the entire voyage.
func IsAnchored(v *Voyage, anchorRadius float64) bool {
	if len(v.Records) < 2 {
		return true
	}
	origin := geo.LatLon{Lat: v.Records[0].Lat, Lon: v.Records[0].Lon}
	for _, r := range v.Records[1:] {
		pt := geo.LatLon{Lat: r.Lat, Lon: r.Lon}
		if geo.Haversine(origin, pt) > anchorRadius {
			return false
		}
	}
	return true
}

// Summary returns a human-readable one-line summary of the voyage.
func (v *Voyage) Summary() string {
	return fmt.Sprintf("voyage %s: %s → %s (%.1f km, %d fixes, max SOG %.1f kn)",
		v.MMSI,
		v.StartTime.Format("2006-01-02T15:04"),
		v.EndTime.Format("2006-01-02T15:04"),
		v.DistanceKm, len(v.Records), v.MaxSOG)
}
