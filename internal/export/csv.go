package export

import (
	"fmt"
	"io"
	"strings"

	"ais-track/internal/detect"
	"ais-track/internal/parse"
)

// WriteTrackCSV writes a vessel's track as CSV with the standard AIS
// header.
func WriteTrackCSV(w io.Writer, track []parse.Record) error {
	if _, err := io.WriteString(w, "mmsi,ts,lat,lon,sog,cog\n"); err != nil {
		return err
	}
	for _, r := range track {
		line := fmt.Sprintf("%s,%s,%.6f,%.6f,%.2f,%.2f\n",
			r.MMSI,
			r.Timestamp.Format("2006-01-02T15:04:05Z"),
			r.Lat, r.Lon, r.SOG, r.COG)
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}

// WriteAnomalyCSV writes detected anomalies as a CSV report.
func WriteAnomalyCSV(w io.Writer, mmsi string, anomalies []detect.Anomaly) error {
	if _, err := io.WriteString(w, "mmsi,kind,timestamp,detail\n"); err != nil {
		return err
	}
	for _, a := range anomalies {
		detail := strings.ReplaceAll(a.Detail, ",", ";")
		line := fmt.Sprintf("%s,%s,%s,%s\n",
			mmsi, a.Kind,
			a.At.Format("2006-01-02T15:04:05Z"),
			detail)
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}

// WriteSummaryCSV writes a per-vessel summary as CSV.
func WriteSummaryCSV(w io.Writer, groups map[string][]parse.Record, maxSOG float64) error {
	if _, err := io.WriteString(w, "mmsi,records,speeding_count,total_anomalies\n"); err != nil {
		return err
	}
	for mmsi, track := range groups {
		anoms := detect.Anomalies(track, maxSOG, nil)
		speed := detect.SpeedingCount(track, maxSOG)
		line := fmt.Sprintf("%s,%d,%d,%d\n", mmsi, len(track), speed, len(anoms))
		if _, err := io.WriteString(w, line); err != nil {
			return err
		}
	}
	return nil
}
