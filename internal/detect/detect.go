// Package detect identifies anomalous vessel behavior from AIS tracks:
// speeding and in-port loitering.
package detect

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// Anomaly describes one detected irregularity on a vessel's track.
type Anomaly struct {
	Kind   string
	At     time.Time
	Detail string
}

// Anomalies inspects a single vessel's track and returns detected anomalies.
//
//   - speeding: one Anomaly per record whose SOG exceeds maxSOG.
//   - loitering: one Anomaly when the vessel is inside the port polygon for
//     >=3 consecutive records.
//
// For a nil or empty track it returns a nil slice.
func Anomalies(track []parse.Record, maxSOG float64, port []geo.LatLon) []Anomaly {
	if len(track) == 0 {
		return nil
	}
	var anomalies []Anomaly

	for _, r := range track {
		if r.SOG > maxSOG {
			anomalies = append(anomalies, Anomaly{
				Kind:   "speeding",
				At:     r.Timestamp,
				Detail: fmt.Sprintf("SOG %.2f exceeds limit %.2f", r.SOG, maxSOG),
			})
		}
	}

	consec := 0
	for _, r := range track {
		inPort := geo.PointInPolygon(geo.LatLon{Lat: r.Lat, Lon: r.Lon}, port)
		if inPort {
			consec++
			if a, ok := loiterTick(consec, r); ok {
				anomalies = append(anomalies, a)
			}
		} else {
			consec = 0
		}
	}
	return anomalies
}

// SpeedingCount returns the number of records in track with SOG > maxSOG.
func SpeedingCount(track []parse.Record, maxSOG float64) int {
	count := 0
	for _, r := range track {
		if r.SOG > maxSOG {
			count++
		}
	}
	return count
}
