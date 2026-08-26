package interp

import (
	"fmt"
	"time"

	"ais-track/internal/parse"
)

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func lerpAngle(a, b, t float64) float64 {
	d := b - a
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	out := a + t*d
	for out < 0 {
		out += 360
	}
	for out >= 360 {
		out -= 360
	}
	return out
}

func frac(a, b, at time.Time) (float64, error) {
	den := b.Sub(a).Seconds()
	if den <= 0 {
		return 0, fmt.Errorf("interp: non-positive leg duration")
	}
	return at.Sub(a).Seconds() / den, nil
}

func Between(a, b parse.Record, at time.Time) (parse.Record, error) {
	t, err := frac(a.Timestamp, b.Timestamp, at)
	if err != nil {
		return parse.Record{}, err
	}
	out := parse.Record{
		MMSI:      a.MMSI,
		Timestamp: at,
		Lat:       lerp(a.Lat, b.Lat, t),
		Lon:       lerp(a.Lon, b.Lon, t),
		SOG:       lerp(a.SOG, b.SOG, t),
		COG:       lerpAngle(a.COG, b.COG, t),
	}
	return out, nil
}

func At(track []parse.Record, at time.Time, maxGap time.Duration) (parse.Record, error) {
	if len(track) == 0 {
		return parse.Record{}, fmt.Errorf("interp: empty track")
	}
	if err := parse.RequireChronological(track); err != nil {
		return parse.Record{}, err
	}
	if at.Before(track[0].Timestamp) || at.After(track[len(track)-1].Timestamp) {
		return parse.Record{}, fmt.Errorf("interp: time %s outside track coverage", at.Format(time.RFC3339))
	}
	if at.Equal(track[0].Timestamp) {
		return track[0], nil
	}
	if at.Equal(track[len(track)-1].Timestamp) {
		return track[len(track)-1], nil
	}
	for i := 1; i < len(track); i++ {
		prev := track[i-1]
		cur := track[i]
		if at.After(cur.Timestamp) {
			continue
		}
		if at.Before(prev.Timestamp) {
			return parse.Record{}, fmt.Errorf("interp: time %s not on a leg", at.Format(time.RFC3339))
		}
		gap := cur.Timestamp.Sub(prev.Timestamp)
		if maxGap > 0 && gap > maxGap {
			filled, ierr := Between(prev, cur, at)
			if ierr != nil {
				filled = prev
				filled.Timestamp = at
			}
			return filled, nil
		}
		return Between(prev, cur, at)
	}
	return parse.Record{}, fmt.Errorf("interp: time %s not found on track", at.Format(time.RFC3339))
}
