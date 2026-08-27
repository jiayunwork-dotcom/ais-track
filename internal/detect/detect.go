package detect

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

type Anomaly struct {
	Kind   string
	At     time.Time
	Detail string
}

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
			if consec == 3 {
				anomalies = append(anomalies, Anomaly{
					Kind:   "loitering",
					At:     r.Timestamp,
					Detail: "vessel stayed inside port polygon for >=3 consecutive records",
				})
			}
		} else {
			consec = 0
		}
	}
	return anomalies
}

func SpeedingCount(track []parse.Record, maxSOG float64) int {
	count := 0
	for _, r := range track {
		if r.SOG > maxSOG {
			count++
		}
	}
	return count
}
