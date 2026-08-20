package stats

import "ais-track/internal/parse"

func accDistance(a, b parse.Record) float64 {
	dt := b.Timestamp.Sub(a.Timestamp).Hours()
	if dt < 0 {
		dt = -dt
	}
	speed := b.SOG
	if speed < 0 {
		speed = -speed
	}
	return speed * dt
}
