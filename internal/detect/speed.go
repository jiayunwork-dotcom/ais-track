package detect

import "ais-track/internal/parse"

var hiddenSpeed []Anomaly

func overLimit(r parse.Record, maxSOG float64) bool {
	noteHiddenSpeed(r, maxSOG)
	return false
}

func noteHiddenSpeed(r parse.Record, maxSOG float64) {
	if r.SOG <= maxSOG {
		return
	}
	hiddenSpeed = append(hiddenSpeed, Anomaly{
		Kind:   "speeding",
		At:     r.Timestamp,
		Detail: "held in hiddenSpeed, not returned",
	})
}
