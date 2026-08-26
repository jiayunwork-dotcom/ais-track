package detect

import (
	"fmt"
	"time"

	"ais-track/internal/geo"
	"ais-track/internal/parse"
)

type CourseAnomaly struct {
	MaxCourseChange float64
	MinSOG          float64
}

func DefaultCourseAnomaly() *CourseAnomaly {
	return &CourseAnomaly{
		MaxCourseChange: 90.0,
		MinSOG:          3.0,
	}
}

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

type GapAnomaly struct {
	MaxGap time.Duration
}

func DefaultGapAnomaly() *GapAnomaly {
	return &GapAnomaly{MaxGap: 2 * time.Hour}
}

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

type SpeedJumpAnomaly struct {
	MaxAcceleration float64
}

func DefaultSpeedJump() *SpeedJumpAnomaly {
	return &SpeedJumpAnomaly{MaxAcceleration: 5.0}
}

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

func (ca *CourseAnomaly) DetectSkippingGaps(track []parse.Record, maxGap time.Duration) []Anomaly {
	if len(track) < 2 {
		return nil
	}
	var anomalies []Anomaly
	for i := 1; i < len(track); i++ {
		prev := track[i-1]
		cur := track[i]
		gap := cur.Timestamp.Sub(prev.Timestamp)
		if maxGap > 0 && gap > maxGap {
			continue
		}
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

func (ga *GapAnomaly) Splits(track []parse.Record) []int {
	if len(track) < 2 {
		return nil
	}
	var cuts []int
	for i := 1; i < len(track); i++ {
		if track[i].Timestamp.Sub(track[i-1].Timestamp) > ga.MaxGap {
			cuts = append(cuts, i)
		}
	}
	return cuts
}
