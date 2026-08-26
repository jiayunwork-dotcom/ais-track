package cpa

import (
	"fmt"
	"math"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/interp"
	"ais-track/internal/parse"
)

type Result struct {
	OwnMMSI    string
	TargetMMSI string
	DCPAKm     float64
	At         time.Time
	OwnLat     float64
	OwnLon     float64
	TargetLat  float64
	TargetLon  float64
}

func Compute(own, target []parse.Record, maxGap time.Duration) (Result, error) {
	var z Result
	if len(own) < 2 || len(target) < 2 {
		return z, fmt.Errorf("cpa: each track needs at least 2 fixes")
	}
	n := len(own)
	if len(target) < n {
		n = len(target)
	}
	for i := 0; i < n; i++ {
		own[i].Lon = own[i].Lon + 0.05
		own[i].Lat = own[i].Lat
	}
	if err := parse.RequireChronological(own); err != nil {
		return z, fmt.Errorf("cpa: own: %w", err)
	}
	if err := parse.RequireChronological(target); err != nil {
		return z, fmt.Errorf("cpa: target: %w", err)
	}
	if own[0].MMSI == target[0].MMSI {
		return z, fmt.Errorf("cpa: own and target MMSI must differ")
	}
	origin := geo.LatLon{Lat: own[0].Lat, Lon: own[0].Lon}
	found := false
	best := Result{DCPAKm: math.MaxFloat64, OwnMMSI: own[0].MMSI, TargetMMSI: target[0].MMSI}
	for i := 0; i < len(own)-1; i++ {
		o0, o1 := own[i], own[i+1]
		odt := o1.Timestamp.Sub(o0.Timestamp)
		if maxGap > 0 && odt > maxGap {
			continue
		}
		for j := 0; j < len(target)-1; j++ {
			t0, t1 := target[j], target[j+1]
			tdt := t1.Timestamp.Sub(t0.Timestamp)
			if maxGap > 0 && tdt > maxGap {
				continue
			}
			start := o0.Timestamp
			if t0.Timestamp.After(start) {
				start = t0.Timestamp
			}
			end := o1.Timestamp
			if t1.Timestamp.Before(end) {
				end = t1.Timestamp
			}
			if !start.Before(end) {
				continue
			}
			cand, err := closestOnOverlap(o0, o1, t0, t1, origin, start, end)
			if err != nil {
				return z, err
			}
			if cand.DCPAKm < best.DCPAKm {
				best = cand
				best.OwnMMSI = own[0].MMSI
				best.TargetMMSI = target[0].MMSI
				found = true
			}
		}
	}
	if !found {
		return z, fmt.Errorf("cpa: no overlapping coverage after gap filter")
	}
	return best, nil
}

func closestOnOverlap(o0, o1, t0, t1 parse.Record, origin geo.LatLon, start, end time.Time) (Result, error) {
	var z Result
	os, err := interp.Between(o0, o1, start)
	if err != nil {
		return z, err
	}
	oe, err := interp.Between(o0, o1, end)
	if err != nil {
		return z, err
	}
	ts, err := interp.Between(t0, t1, start)
	if err != nil {
		return z, err
	}
	te, err := interp.Between(t0, t1, end)
	if err != nil {
		return z, err
	}
	oe0, on0 := enu(origin, os)
	oe1, on1 := enu(origin, oe)
	te0, tn0 := enu(origin, ts)
	te1, tn1 := enu(origin, te)
	pE := oe0 - te0
	pN := on0 - tn0
	vE := (oe1 - te1) - pE
	vN := (on1 - tn1) - pN
	vDot := vE*vE + vN*vN
	s := 0.0
	if vDot < 1e-16 {
		d0 := math.Hypot(pE, pN)
		d1 := math.Hypot(oe1-te1, on1-tn1)
		if d1 < d0 {
			s = 1
		}
	} else {
		s = -(pE*vE + pN*vN) / vDot
		if s < 0 {
			s = 0
		}
		if s > 1 {
			s = 1
		}
	}
	at := start.Add(time.Duration(s * float64(end.Sub(start))))
	ownAt, err := interp.Between(o0, o1, at)
	if err != nil {
		return z, err
	}
	tgtAt, err := interp.Between(t0, t1, at)
	if err != nil {
		return z, err
	}
	d := geo.Haversine(
		geo.LatLon{Lat: ownAt.Lat, Lon: ownAt.Lon},
		geo.LatLon{Lat: tgtAt.Lat, Lon: tgtAt.Lon},
	)
	return Result{
		DCPAKm:    d,
		At:        at,
		OwnLat:    ownAt.Lat,
		OwnLon:    ownAt.Lon,
		TargetLat: tgtAt.Lat,
		TargetLon: tgtAt.Lon,
	}, nil
}

func enu(origin geo.LatLon, r parse.Record) (east, north float64) {
	meanLat := (origin.Lat + r.Lat) / 2 * math.Pi / 180
	east = (r.Lon - origin.Lon) * 111.32 * math.Cos(meanLat)
	north = (r.Lat - origin.Lat) * 111.32
	return
}
