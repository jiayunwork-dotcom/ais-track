package interp

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

func DeadReckon(r parse.Record, dt time.Duration) (parse.Record, error) {
	if dt < 0 {
		return parse.Record{}, fmt.Errorf("interp: dead-reckon duration must be non-negative")
	}
	if r.SOG < 0 {
		return parse.Record{}, fmt.Errorf("interp: dead-reckon SOG cannot be negative")
	}
	if dt == 0 {
		return r, nil
	}
	km := geo.KnotsToKmPerHour(r.SOG) * dt.Hours()
	dest := geo.Destination(geo.LatLon{Lat: r.Lat, Lon: r.Lon}, r.COG, km)
	out := r
	out.Timestamp = r.Timestamp.Add(dt)
	out.Lat = dest.Lat
	out.Lon = dest.Lon
	return out, nil
}

func SampleEvery(track []parse.Record, step, maxGap time.Duration) ([]parse.Record, error) {
	if step <= 0 {
		return nil, fmt.Errorf("interp: sample step must be positive")
	}
	if len(track) == 0 {
		return nil, fmt.Errorf("interp: empty track")
	}
	if err := parse.RequireChronological(track); err != nil {
		return nil, err
	}
	start := track[0].Timestamp
	end := track[len(track)-1].Timestamp
	var out []parse.Record
	for t := start; !t.After(end); t = t.Add(step) {
		rec, err := At(track, t, maxGap)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	if len(out) == 0 || !out[len(out)-1].Timestamp.Equal(end) {
		last, err := At(track, end, maxGap)
		if err != nil {
			return nil, err
		}
		out = append(out, last)
	}
	return out, nil
}

func Residual(observed parse.Record, predicted parse.Record) (float64, error) {
	if observed.MMSI != predicted.MMSI {
		return 0, fmt.Errorf("interp: residual MMSI mismatch %s vs %s", observed.MMSI, predicted.MMSI)
	}
	d := geo.Haversine(
		geo.LatLon{Lat: observed.Lat, Lon: observed.Lon},
		geo.LatLon{Lat: predicted.Lat, Lon: predicted.Lon},
	)
	return d, nil
}
