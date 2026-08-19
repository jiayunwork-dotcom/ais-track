// Package stats computes per-vessel and fleet-wide statistical summaries
// from AIS track data: distance traveled, time underway, speed
// distributions, and reporting regularity.
package stats

import (
	"math"
	"sort"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// VesselStats holds aggregated statistics for a single vessel's track.
type VesselStats struct {
	MMSI           string
	RecordCount    int
	TotalDistKm    float64
	TotalDuration  time.Duration
	MeanSOG        float64
	MedianSOG      float64
	MaxSOG         float64
	MinSOG         float64
	MeanGapSec     float64
	MaxGapSec      float64
	SpeedVariance  float64
}

// Compute calculates statistics for a single vessel's sorted track.
func Compute(track []parse.Record) *VesselStats {
	if len(track) == 0 {
		return &VesselStats{}
	}
	s := &VesselStats{
		MMSI:        track[0].MMSI,
		RecordCount: len(track),
		MinSOG:      math.MaxFloat64,
	}

	speeds := make([]float64, 0, len(track))
	var sumSOG float64

	for i, r := range track {
		speeds = append(speeds, r.SOG)
		sumSOG += r.SOG
		if r.SOG > s.MaxSOG {
			s.MaxSOG = r.SOG
		}
		if r.SOG < s.MinSOG {
			s.MinSOG = r.SOG
		}
		if i > 0 {
			s.TotalDistKm += geo.LegDistance(track[i-1], r)
		}
	}

	s.MeanSOG = sumSOG / float64(len(track))
	s.MedianSOG = median(speeds)
	s.SpeedVariance = variance(speeds, s.MeanSOG)

	// Gaps between consecutive records
	if len(track) > 1 {
		s.TotalDuration = track[len(track)-1].Timestamp.Sub(track[0].Timestamp)
		var sumGap float64
		for i := 1; i < len(track); i++ {
			gap := track[i].Timestamp.Sub(track[i-1].Timestamp).Seconds()
			sumGap += gap
			if gap > s.MaxGapSec {
				s.MaxGapSec = gap
			}
		}
		s.MeanGapSec = sumGap / float64(len(track)-1)
	}
	return s
}

// FleetStats aggregates statistics across multiple vessels.
type FleetStats struct {
	VesselCount   int
	TotalRecords  int
	TotalDistKm   float64
	MeanDistKm    float64
	MaxDistKm     float64
	MeanDuration  time.Duration
	MeanMeanSOG   float64
}

// ComputeFleet calculates fleet-wide statistics from per-vessel stats.
func ComputeFleet(vessels []*VesselStats) *FleetStats {
	if len(vessels) == 0 {
		return &FleetStats{}
	}
	fs := &FleetStats{VesselCount: len(vessels)}
	var sumDist, sumMeanSOG float64
	var sumDur time.Duration

	for _, v := range vessels {
		fs.TotalRecords += v.RecordCount
		fs.TotalDistKm += v.TotalDistKm
		sumDist += v.TotalDistKm
		sumMeanSOG += v.MeanSOG
		sumDur += v.TotalDuration
		if v.TotalDistKm > fs.MaxDistKm {
			fs.MaxDistKm = v.TotalDistKm
		}
	}
	fs.MeanDistKm = sumDist / float64(len(vessels))
	fs.MeanMeanSOG = sumMeanSOG / float64(len(vessels))
	fs.MeanDuration = sumDur / time.Duration(len(vessels))
	return fs
}

func median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func variance(vals []float64, mean float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return sum / float64(len(vals)-1)
}
