package report

import (
	"fmt"
	"io"
	"strings"

	"ais-track/internal/detect"
	"ais-track/internal/parse"
	"ais-track/internal/stats"
)

type VesselReport struct {
	MMSI        string
	Stats       *stats.VesselStats
	Anomalies   []detect.Anomaly
	RecordCount int
}

type FleetReport struct {
	Vessels    []VesselReport
	Fleet      *stats.FleetStats
	TotalAnoms int
}

func Generate(groups map[string][]parse.Record, maxSOG float64, port []detect.Zone) *FleetReport {
	fr := &FleetReport{}
	var vesselStats []*stats.VesselStats

	for mmsi, track := range groups {
		vs := stats.Compute(track)
		vesselStats = append(vesselStats, vs)

		var allAnoms []detect.Anomaly

		anoms := detect.Anomalies(track, maxSOG, nil)
		allAnoms = append(allAnoms, anoms...)

		ca := detect.DefaultCourseAnomaly()
		allAnoms = append(allAnoms, ca.Detect(track)...)

		ga := detect.DefaultGapAnomaly()
		allAnoms = append(allAnoms, ga.Detect(track)...)

		sj := detect.DefaultSpeedJump()
		allAnoms = append(allAnoms, sj.Detect(track)...)

		for _, z := range port {
			violations := detect.ZoneViolation(track, []detect.Zone{z})
			allAnoms = append(allAnoms, violations...)
		}

		fr.Vessels = append(fr.Vessels, VesselReport{
			MMSI:        mmsi,
			Stats:       vs,
			Anomalies:   allAnoms,
			RecordCount: len(track),
		})
		fr.TotalAnoms += len(allAnoms)
	}

	fr.Fleet = stats.ComputeFleet(vesselStats)
	return fr
}

func WriteText(w io.Writer, fr *FleetReport) error {
	header := fmt.Sprintf("Fleet Report: %d vessels, %d total records, %d anomalies\n",
		fr.Fleet.VesselCount, fr.Fleet.TotalRecords, fr.TotalAnoms)
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	if _, err := io.WriteString(w, strings.Repeat("=", 60)+"\n\n"); err != nil {
		return err
	}

	for _, vr := range fr.Vessels {
		line := fmt.Sprintf("Vessel %s: %d records, %.1f km, avg SOG %.1f kn, %d anomalies\n",
			vr.MMSI, vr.RecordCount, vr.Stats.TotalDistKm,
			vr.Stats.MeanSOG, len(vr.Anomalies))
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
		for _, a := range vr.Anomalies {
			detail := fmt.Sprintf("  [%s] %s: %s\n",
				a.Kind, a.At.Format("2006-01-02T15:04"), a.Detail)
			if _, err := io.WriteString(w, detail); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

func CountByKind(fr *FleetReport) map[string]int {
	counts := make(map[string]int)
	for _, vr := range fr.Vessels {
		for _, a := range vr.Anomalies {
			counts[a.Kind]++
		}
	}
	return counts
}
