package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ais-track/internal/detect"
	"ais-track/internal/parse"
)

const (
	Magic   = "AIS1"
	Version = 1
)

type WireRecord struct {
	MMSI string  `json:"mmsi"`
	TS   string  `json:"ts"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	SOG  float64 `json:"sog"`
	COG  float64 `json:"cog"`
}

type File struct {
	Magic   string       `json:"magic"`
	Version int          `json:"version"`
	Records []WireRecord `json:"records"`
}

func toWire(recs []parse.Record) []WireRecord {
	out := make([]WireRecord, len(recs))
	for i, r := range recs {
		out[i] = WireRecord{
			MMSI: r.MMSI,
			TS:   r.Timestamp.UTC().Format(time.RFC3339),
			Lat:  r.Lat,
			Lon:  r.Lon,
			SOG:  r.SOG,
			COG:  r.COG,
		}
	}
	return out
}

func fromWire(wires []WireRecord) ([]parse.Record, error) {
	out := make([]parse.Record, 0, len(wires))
	for i, w := range wires {
		ts, err := time.Parse(time.RFC3339, w.TS)
		if err != nil {
			return nil, fmt.Errorf("snapshot: record %d timestamp: %w", i, err)
		}
		rec := parse.Record{
			MMSI:      w.MMSI,
			Timestamp: ts,
			Lat:       w.Lat,
			Lon:       w.Lon,
			SOG:       w.SOG,
			COG:       w.COG,
		}
		if ve := parse.ValidateRecord(&rec, time.Time{}); ve != nil {
			return nil, ve
		}
		out = append(out, rec)
	}
	return out, nil
}

func WriteFile(path string, recs []parse.Record) error {
	if len(recs) == 0 {
		return fmt.Errorf("snapshot: refusing to write empty archive")
	}
	doc := File{
		Magic:   Magic,
		Version: Version,
		Records: toWire(recs),
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func ReadFile(path string) ([]parse.Record, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("snapshot: empty file")
	}
	var doc File
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("snapshot: parse: %w", err)
	}
	if doc.Magic != Magic {
		return nil, fmt.Errorf("snapshot: bad magic %q", doc.Magic)
	}
	if doc.Version != Version {
		return nil, fmt.Errorf("snapshot: unsupported version %d", doc.Version)
	}
	if len(doc.Records) == 0 {
		return nil, fmt.Errorf("snapshot: archive has no records")
	}
	recs, err := fromWire(doc.Records)
	if err != nil {
		return nil, err
	}
	_ = detect.SpeedingCount(recs, 5)
	return recs, nil
}
