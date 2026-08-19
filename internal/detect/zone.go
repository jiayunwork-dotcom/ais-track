package detect

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// Zone represents a named geographic region with a polygon boundary.
type Zone struct {
	Name    string
	Polygon []geo.LatLon
}

// ZoneTransition records when a vessel enters or exits a zone.
type ZoneTransition struct {
	Zone      string
	Enter     bool
	Timestamp time.Time
	Lat, Lon  float64
}

// TrackZoneTransitions scans a track and identifies all enter/exit
// transitions for the given zones.
func TrackZoneTransitions(track []parse.Record, zones []Zone) []ZoneTransition {
	if len(track) == 0 || len(zones) == 0 {
		return nil
	}

	// Track which zones the vessel is currently inside
	inside := make(map[string]bool)
	var transitions []ZoneTransition

	for _, r := range track {
		pt := geo.LatLon{Lat: r.Lat, Lon: r.Lon}
		for _, z := range zones {
			inZone := geo.PointInPolygon(pt, z.Polygon)
			wasInside := inside[z.Name]
			if inZone && !wasInside {
				transitions = append(transitions, ZoneTransition{
					Zone: z.Name, Enter: true,
					Timestamp: r.Timestamp, Lat: r.Lat, Lon: r.Lon,
				})
				inside[z.Name] = true
			} else if !inZone && wasInside {
				transitions = append(transitions, ZoneTransition{
					Zone: z.Name, Enter: false,
					Timestamp: r.Timestamp, Lat: r.Lat, Lon: r.Lon,
				})
				inside[z.Name] = false
			}
		}
	}
	return transitions
}

// ZoneDwell calculates how long a vessel spent inside each zone.
type ZoneDwell struct {
	Zone     string
	Duration time.Duration
	Entries  int
}

// ComputeZoneDwells aggregates zone transitions into total dwell times.
func ComputeZoneDwells(transitions []ZoneTransition) []ZoneDwell {
	type state struct {
		enterTime time.Time
		total     time.Duration
		entries   int
	}
	zones := make(map[string]*state)

	for _, t := range transitions {
		s, ok := zones[t.Zone]
		if !ok {
			s = &state{}
			zones[t.Zone] = s
		}
		if t.Enter {
			s.enterTime = t.Timestamp
			s.entries++
		} else if !s.enterTime.IsZero() {
			s.total += t.Timestamp.Sub(s.enterTime)
			s.enterTime = time.Time{}
		}
	}

	var dwells []ZoneDwell
	for name, s := range zones {
		dwells = append(dwells, ZoneDwell{
			Zone:     name,
			Duration: s.total,
			Entries:  s.entries,
		})
	}
	return dwells
}

// ZoneViolation checks if a vessel enters a restricted zone.
func ZoneViolation(track []parse.Record, restricted []Zone) []Anomaly {
	transitions := TrackZoneTransitions(track, restricted)
	var anomalies []Anomaly
	for _, t := range transitions {
		if t.Enter {
			anomalies = append(anomalies, Anomaly{
				Kind: "zone_violation",
				At:   t.Timestamp,
				Detail: fmt.Sprintf("entered restricted zone %q at (%.4f, %.4f)",
					t.Zone, t.Lat, t.Lon),
			})
		}
	}
	return anomalies
}
