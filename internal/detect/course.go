package detect

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

// CourseAnomaly detects sudden course changes that may indicate evasive
// maneuvering or navigation system issues.
type CourseAnomaly struct {
	// MaxCourseChange is the threshold in degrees for a single-step
	// course change to be flagged.
	MaxCourseChange float64
	// MinSOG is the minimum speed for the course change to be considered
	// meaningful (low-speed vessels may show erratic headings).
	MinSOG float64
}

// DefaultCourseAnomaly returns sensible defaults for course anomaly
// detection.
func DefaultCourseAnomaly() *CourseAnomaly {
	return &CourseAnomaly{
		MaxCourseChange: 90.0,
		MinSOG:          3.0,
	}
}

// Detect scans a track for sudden course changes and returns anomalies.
func (ca *CourseAnomaly) Detect(track []parse.Record) []Anomaly {
	if len(track) < 2 {
		return nil
	}
	var anomalies []Anomaly
	for i := 1; i < len(track); i++ {
		prev := track[i-1]
		cur := track[i]
		if cur.SOG < ca.MinSOG {
			continue
		}
		change := geo.CourseChange(prev.COG, cur.COG)
		if change > ca.MaxCourseChange {
			anomalies = append(anomalies, Anomaly{
				Kind: "course_change",
				At:   cur.Timestamp,
				Detail: fmt.Sprintf("COG changed %.1f degrees (%.1f -> %.1f) at SOG %.1f",
					change, prev.COG, cur.COG, cur.SOG),
			})
		}
	}
	return anomalies
}

// GapAnomaly detects unreasonably large time gaps in AIS reporting which
// may indicate intentional AIS shutdown (dark vessel behavior).
type GapAnomaly struct {
	MaxGap time.Duration
}

// DefaultGapAnomaly returns a default gap detector with a 2-hour threshold.
func DefaultGapAnomaly() *GapAnomaly {
	return &GapAnomaly{MaxGap: 2 * time.Hour}
}

// Detect scans a track for reporting gaps exceeding the threshold.
func (ga *GapAnomaly) Detect(track []parse.Record) []Anomaly {
	if len(track) < 2 {
		return nil
	}
	var anomalies []Anomaly
	for i := 1; i < len(track); i++ {
		gap := track[i].Timestamp.Sub(track[i-1].Timestamp)
		if gap > ga.MaxGap {
			anomalies = append(anomalies, Anomaly{
				Kind: "reporting_gap",
				At:   track[i].Timestamp,
				Detail: fmt.Sprintf("gap of %v between records (threshold %v)",
					gap.Round(time.Minute), ga.MaxGap),
			})
		}
	}
	return anomalies
}

// SpeedJumpAnomaly detects unrealistic speed jumps between consecutive
// records that may indicate position spoofing.
type SpeedJumpAnomaly struct {
	MaxAcceleration float64 // knots per minute
}

// DefaultSpeedJump returns a default speed jump detector.
func DefaultSpeedJump() *SpeedJumpAnomaly {
	return &SpeedJumpAnomaly{MaxAcceleration: 5.0}
}

// Detect scans for unrealistic speed changes.
func (sj *SpeedJumpAnomaly) Detect(track []parse.Record) []Anomaly {
	if len(track) < 2 {
		return nil
	}
	var anomalies []Anomaly
	for i := 1; i < len(track); i++ {
		dt := track[i].Timestamp.Sub(track[i-1].Timestamp).Minutes()
		if dt <= 0 {
			continue
		}
		dSOG := track[i].SOG - track[i-1].SOG
		if dSOG < 0 {
			dSOG = -dSOG
		}
		accel := dSOG / dt
		if accel > sj.MaxAcceleration {
			anomalies = append(anomalies, Anomaly{
				Kind: "speed_jump",
				At:   track[i].Timestamp,
				Detail: fmt.Sprintf("acceleration %.2f kn/min exceeds limit %.2f",
					accel, sj.MaxAcceleration),
			})
		}
	}
	return anomalies
}
