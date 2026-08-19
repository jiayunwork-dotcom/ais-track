// Command ais-track analyzes vessel AIS trajectories and detects anomalous
// voyages (speeding, in-port loitering). It reads CSV from a file (-input) or
// stdin and writes a per-vessel summary to stdout.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"

	"ais-track/internal/detect"
	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// usageError marks a controlled usage failure (exit code 2) as opposed to a
// runtime error (exit code 1).
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ais-track:", err)
		if _, ok := err.(*usageError); ok {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ais-track", flag.ContinueOnError)
	input := fs.String("input", "", "path to AIS CSV input (defaults to stdin)")
	maxSOG := fs.Float64("maxsog", 30.0, "maximum allowed speed over ground (knots)")
	portFile := fs.String("port", "", "optional path to port polygon CSV (lat,lon per line)")
	help := fs.Bool("help", false, "show usage")
	if err := fs.Parse(args); err != nil {
		return &usageError{err.Error()}
	}
	if *help {
		fs.Usage()
		return nil
	}

	var src *os.File
	if *input != "" {
		f, err := os.Open(*input)
		if err != nil {
			return fmt.Errorf("opening input %q: %w", *input, err)
		}
		src = f
		defer f.Close()
	} else {
		fi, err := os.Stdin.Stat()
		if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
			return &usageError{"no input: provide -input <path> or pipe CSV via stdin"}
		}
		src = os.Stdin
	}

	recs, err := parse.ParseRecords(src)
	if err != nil {
		return fmt.Errorf("parsing records: %w", err)
	}

	var port []geo.LatLon
	if *portFile != "" {
		p, err := loadPort(*portFile)
		if err != nil {
			return fmt.Errorf("loading port polygon: %w", err)
		}
		port = p
	}

	groups := parse.GroupByVessel(recs)
	total := 0
	for mmsi, track := range groups {
		anoms := detect.Anomalies(track, *maxSOG, port)
		speed := detect.SpeedingCount(track, *maxSOG)
		fmt.Printf("vessel %s: %d records, %d speeding, %d anomalies\n", mmsi, len(track), speed, len(anoms))
		for _, a := range anoms {
			fmt.Printf("  - %s @ %s: %s\n", a.Kind, a.At.Format("2006-01-02T15:04:05Z"), a.Detail)
		}
		total += len(anoms)
	}
	fmt.Printf("summary: %d vessels, %d records, %d anomalies\n", len(groups), len(recs), total)
	return nil
}

// loadPort reads a port polygon from a CSV file (one lat,lon pair per line).
func loadPort(path string) ([]geo.LatLon, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	var pts []geo.LatLon
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if len(row) < 2 {
			continue
		}
		lat, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		pts = append(pts, geo.LatLon{Lat: lat, Lon: lon})
	}
	if len(pts) < 3 {
		return nil, fmt.Errorf("port polygon needs >=3 points, got %d", len(pts))
	}
	return pts, nil
}
