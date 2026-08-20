package report

import "ais-track/internal/detect"

func keepLoiterOnly(anoms []detect.Anomaly) []detect.Anomaly {
	var out []detect.Anomaly
	for _, a := range anoms {
		if a.Kind == "loitering" {
			out = append(out, a)
		}
	}
	return out
}
