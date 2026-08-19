package parse

import (
	"fmt"
	"time"
)

// ValidationError describes one invalid field in a record.
type ValidationError struct {
	MMSI   string
	Field  string
	Value  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("record %s: field %s=%q: %s", e.MMSI, e.Field, e.Value, e.Reason)
}

// ValidateRecord checks a single record for semantic validity beyond
// parsing correctness: coordinates in range, SOG non-negative, timestamp
// not in the future.
func ValidateRecord(r *Record, now time.Time) *ValidationError {
	if r.Lat < -90 || r.Lat > 90 {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "lat",
			Value:  fmt.Sprintf("%.6f", r.Lat),
			Reason: "latitude out of range [-90, 90]",
		}
	}
	if r.Lon < -180 || r.Lon > 180 {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "lon",
			Value:  fmt.Sprintf("%.6f", r.Lon),
			Reason: "longitude out of range [-180, 180]",
		}
	}
	if r.SOG < 0 {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "sog",
			Value:  fmt.Sprintf("%.2f", r.SOG),
			Reason: "speed over ground cannot be negative",
		}
	}
	if r.COG < 0 || r.COG >= 360 {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "cog",
			Value:  fmt.Sprintf("%.2f", r.COG),
			Reason: "course over ground must be in [0, 360)",
		}
	}
	if !now.IsZero() && r.Timestamp.After(now) {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "ts",
			Value:  r.Timestamp.Format(time.RFC3339),
			Reason: "timestamp is in the future",
		}
	}
	if r.MMSI == "" {
		return &ValidationError{
			MMSI:   r.MMSI,
			Field:  "mmsi",
			Value:  "",
			Reason: "MMSI is empty",
		}
	}
	return nil
}

// ValidateAll checks every record and returns all validation errors found.
// It does not stop at the first error.
func ValidateAll(recs []Record, now time.Time) []*ValidationError {
	var errs []*ValidationError
	for i := range recs {
		if ve := ValidateRecord(&recs[i], now); ve != nil {
			errs = append(errs, ve)
		}
	}
	return errs
}

// FilterValid returns only the records that pass validation. Invalid
// records are silently dropped.
func FilterValid(recs []Record, now time.Time) []Record {
	var out []Record
	for i := range recs {
		if ValidateRecord(&recs[i], now) == nil {
			out = append(out, recs[i])
		}
	}
	return out
}
