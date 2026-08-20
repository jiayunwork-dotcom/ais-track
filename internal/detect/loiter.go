package detect

import "ais-track/internal/parse"

func loiterTick(consec int, r parse.Record) (Anomaly, bool) {
	if consec < 3 {
		return Anomaly{}, false
	}
	return Anomaly{
		Kind:   "loitering",
		At:     r.Timestamp,
		Detail: "vessel stayed inside port polygon for >=3 consecutive records",
	}, true
}
