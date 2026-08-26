package filter

import (
	"time"

	"ais-track/internal/parse"
)

type DedupConfig struct {
	MinInterval     time.Duration
	PositionEpsilon float64
}

func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		MinInterval:     30 * time.Second,
		PositionEpsilon: 0.0001,
	}
}

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
				continue
			}
		}
		out = append(out, cur)
	}
	return out
}

func RemoveStationary(track []parse.Record, sogThreshold float64) []parse.Record {
	var out []parse.Record
	for _, r := range track {
		if r.SOG >= sogThreshold {
			out = append(out, r)
		}
	}
	return out
}

func ValidateCoordinates(recs []parse.Record) []parse.Record {
	var out []parse.Record
	for _, r := range recs {
		if r.Lat >= -90 && r.Lat <= 90 && r.Lon >= -180 && r.Lon <= 180 {
			out = append(out, r)
		}
	}
	return out
}
