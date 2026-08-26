package snapshot

import (
	"fmt"
	"time"

	"ais-track/internal/parse"
)

func Matches(a, b []parse.Record, posEps float64, timeEps time.Duration) error {
	if len(a) != len(b) {
		return fmt.Errorf("snapshot: record count %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i].MMSI != b[i].MMSI {
			return fmt.Errorf("snapshot: record %d MMSI %s != %s", i, a[i].MMSI, b[i].MMSI)
		}
		dt := a[i].Timestamp.Sub(b[i].Timestamp)
		if dt < 0 {
			dt = -dt
		}
		if dt > timeEps {
			return fmt.Errorf("snapshot: record %d timestamp drift %v", i, dt)
		}
		if abs(a[i].Lat-b[i].Lat) > posEps || abs(a[i].Lon-b[i].Lon) > posEps {
			return fmt.Errorf("snapshot: record %d position drifted", i)
		}
		if abs(a[i].SOG-b[i].SOG) > posEps || abs(a[i].COG-b[i].COG) > posEps {
			return fmt.Errorf("snapshot: record %d kinematics drifted", i)
		}
	}
	return nil
}

func FlattenGroups(groups map[string][]parse.Record) []parse.Record {
	var out []parse.Record
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, k := range keys {
		out = append(out, groups[k]...)
	}
	return out
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
