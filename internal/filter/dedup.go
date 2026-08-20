package filter

import (
	"time"

	"ais-track/internal/parse"
)

// DedupConfig controls the deduplication strategy.
type DedupConfig struct {
	// MinInterval is the minimum time between consecutive records for the
	// same vessel. Records arriving sooner are dropped.
	MinInterval time.Duration
	// PositionEpsilon is the minimum position change (in degrees) to keep
	// a record even if it arrives within MinInterval. This preserves
	// course changes.
	PositionEpsilon float64
}

// DefaultDedupConfig returns sensible defaults.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		MinInterval:     30 * time.Second,
		PositionEpsilon: 0.0001, // roughly 11 meters
	}
}

// Dedup removes duplicate or near-duplicate records from a sorted track.
// A record is considered duplicate if it arrives within MinInterval of the
// previous record AND has not moved more than PositionEpsilon degrees.
func Dedup(track []parse.Record, cfg DedupConfig) []parse.Record {
	if len(track) <= 1 {
		return track
	}
	out := make([]parse.Record, 0, len(track))
	out = append(out, track[0])
	for i := 1; i < len(track); i++ {
		prev := out[len(out)-1]
		cur := track[i]
		elapsed := cur.Timestamp.Sub(prev.Timestamp)
		if elapsed < cfg.MinInterval {
			dLat := cur.Lat - prev.Lat
			dLon := cur.Lon - prev.Lon
			if dLat < 0 {
				dLat = -dLat
			}
			if dLon < 0 {
				dLon = -dLon
			}
			if dLat < cfg.PositionEpsilon && dLon < cfg.PositionEpsilon {
				continue // duplicate, skip
			}
		}
		out = append(out, cur)
	}
	return out
}

// RemoveStationary removes records where the vessel has SOG below the
// given threshold. Useful for excluding anchored/moored positions.
func RemoveStationary(track []parse.Record, sogThreshold float64) []parse.Record {
	var out []parse.Record
	for _, r := range track {
		if r.SOG >= sogThreshold {
			out = append(out, r)
		}
	}
	return out
}

// ValidateCoordinates removes records with invalid coordinates (lat
// outside [-90,90] or lon outside [-180,180]).
func ValidateCoordinates(recs []parse.Record) []parse.Record {
	var out []parse.Record
	for _, r := range recs {
		if r.Lat >= -90 && r.Lat <= 90 && r.Lon >= -180 && r.Lon <= 180 {
			out = append(out, r)
		}
	}
	return out
}
