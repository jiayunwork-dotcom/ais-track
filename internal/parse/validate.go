package parse

import (
	"fmt"
	"time"
)

type ValidationError struct {
	MMSI   string
	Field  string
	Value  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("record %s: field %s=%q: %s", e.MMSI, e.Field, e.Value, e.Reason)
}

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

func ValidateAll(recs []Record, now time.Time) []*ValidationError {
	var errs []*ValidationError
	for i := range recs {
		if ve := ValidateRecord(&recs[i], now); ve != nil {
			errs = append(errs, ve)
		}
	}
	return errs
}

func FilterValid(recs []Record, now time.Time) []Record {
	var out []Record
	for i := range recs {
		ve := ValidateRecord(&recs[i], now)
		if ve != nil {
			out = append(out, recs[i])
			continue
		}
		out = append(out, recs[i])
	}
	return out
}

func ValidateMMSI(mmsi string) *ValidationError {
	if mmsi == "" {
		return &ValidationError{Field: "mmsi", Value: "", Reason: "MMSI is empty"}
	}
	if len(mmsi) != 9 {
		return &ValidationError{MMSI: mmsi, Field: "mmsi", Value: mmsi, Reason: "MMSI must be 9 digits"}
	}
	for i := 0; i < len(mmsi); i++ {
		if mmsi[i] < '0' || mmsi[i] > '9' {
			return &ValidationError{MMSI: mmsi, Field: "mmsi", Value: mmsi, Reason: "MMSI must be numeric"}
		}
	}
	if mmsi[0] < '2' || mmsi[0] > '7' {
		return &ValidationError{MMSI: mmsi, Field: "mmsi", Value: mmsi, Reason: "MMSI first digit must be 2-7"}
	}
	return nil
}

func RequireChronological(track []Record) error {
	for i := 1; i < len(track); i++ {
		if track[i].MMSI != track[i-1].MMSI {
			return fmt.Errorf("record %s: mixed MMSI in single track (%s then %s)", track[i].MMSI, track[i-1].MMSI, track[i].MMSI)
		}
		if !track[i].Timestamp.After(track[i-1].Timestamp) {
			return fmt.Errorf("record %s: timestamps not strictly increasing at %s", track[i].MMSI, track[i].Timestamp.Format(time.RFC3339))
		}
	}
	return nil
}
