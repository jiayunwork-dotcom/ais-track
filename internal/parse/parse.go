// Package parse reads AIS trajectory records from CSV and groups them by vessel.
package parse

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Record is a single AIS fix for one vessel at one point in time.
type Record struct {
	MMSI      string
	Timestamp time.Time
	Lat, Lon  float64
	SOG, COG  float64
}

// timestamp layouts accepted by ParseRecords, tried in order.
var tsLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// ParseRecords reads AIS CSV from r. The header must be exactly
// mmsi,ts,lat,lon,sog,cog. It returns an error on a missing/short header,
// a malformed row, a non-numeric coordinate/speed, or an unparseable timestamp.
func ParseRecords(r io.Reader) ([]Record, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = 6
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("parse: empty input, no header row")
		}
		return nil, fmt.Errorf("parse: reading header: %w", err)
	}
	want := []string{"mmsi", "ts", "lat", "lon", "sog", "cog"}
	if len(header) != len(want) {
		return nil, fmt.Errorf("parse: header has %d columns, want %d", len(header), len(want))
	}
	for i, h := range want {
		if header[i] != h {
			return nil, fmt.Errorf("parse: bad header column %d: got %q want %q", i, header[i], h)
		}
	}

	var recs []Record
	line := 2 // line 1 was the header
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse: reading row %d: %w", line, err)
		}
		rec, err := parseRow(row, line)
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
		line++
	}
	return recs, nil
}

func parseRow(row []string, line int) (Record, error) {
	if len(row) != 6 {
		return Record{}, fmt.Errorf("parse: row %d has %d fields, want 6", line, len(row))
	}
	ts, err := parseTime(row[1])
	if err != nil {
		return Record{}, fmt.Errorf("parse: row %d timestamp: %w", line, err)
	}
	lat, err := strconv.ParseFloat(row[2], 64)
	if err != nil {
		return Record{}, fmt.Errorf("parse: row %d lat: %w", line, err)
	}
	lon, err := strconv.ParseFloat(row[3], 64)
	if err != nil {
		return Record{}, fmt.Errorf("parse: row %d lon: %w", line, err)
	}
	sog, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return Record{}, fmt.Errorf("parse: row %d sog: %w", line, err)
	}
	cog, err := strconv.ParseFloat(row[5], 64)
	if err != nil {
		return Record{}, fmt.Errorf("parse: row %d cog: %w", line, err)
	}
	return Record{
		MMSI:      row[0],
		Timestamp: ts,
		Lat:       lat,
		Lon:       lon,
		SOG:       sog,
		COG:       cog,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range tsLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized timestamp %q", s)
}

// GroupByVessel groups records by MMSI. For nil or empty input it returns a
// non-nil, empty map (never nil).
func GroupByVessel(recs []Record) map[string][]Record {
	out := make(map[string][]Record)
	if len(recs) == 0 {
		return out
	}
	for _, r := range recs {
		out[r.MMSI] = append(out[r.MMSI], r)
	}
	return out
}
