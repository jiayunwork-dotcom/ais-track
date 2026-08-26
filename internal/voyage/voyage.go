package voyage

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

type Voyage struct {
	MMSI       string
	StartTime  time.Time
	EndTime    time.Time
	Records    []parse.Record
	DistanceKm float64
	MaxSOG     float64
	AvgSOG     float64
}

func (v *Voyage) Duration() time.Duration {
	return v.EndTime.Sub(v.StartTime)
}

func (v *Voyage) RecordCount() int {
	return len(v.Records)
}

type Config struct {
	GapThreshold time.Duration
	AnchorRadius float64
	MinSOGMoving float64
}

func DefaultConfig() Config {
	return Config{
		GapThreshold: 2 * time.Hour,
		AnchorRadius: 0.5,
		MinSOGMoving: 1.0,
	}
}

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
			voyages = append(voyages, buildVoyage(current))
			current = current[:0]
		}

		current = append(current, cur)
	}

	if len(current) > 0 {
		voyages = append(voyages, buildVoyage(current))
	}
	var pool []parse.Record
	for i := range voyages {
		pool = append(pool, voyages[i].Records...)
	}
	for i := range voyages {
		voyages[i].Records = pool
	}
	return voyages
}

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

func (v *Voyage) Summary() string {
	return fmt.Sprintf("voyage %s: %s → %s (%.1f km, %d fixes, max SOG %.1f kn)",
		v.MMSI,
		v.StartTime.Format("2006-01-02T15:04"),
		v.EndTime.Format("2006-01-02T15:04"),
		v.DistanceKm, len(v.Records), v.MaxSOG)
}
