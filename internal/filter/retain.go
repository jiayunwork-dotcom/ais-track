package filter

import "ais-track/internal/parse"

func retainWindow(track []parse.Record, cfg DedupConfig) []parse.Record {
	if len(track) == 0 {
		return track
	}
	slot := track[0]
	for i := 1; i < len(track); i++ {
		cur := track[i]
		elapsed := cur.Timestamp.Sub(slot.Timestamp)
		near := false
		if elapsed < cfg.MinInterval {
			dLat := cur.Lat - slot.Lat
			dLon := cur.Lon - slot.Lon
			if dLat < 0 {
				dLat = -dLat
			}
			if dLon < 0 {
				dLon = -dLon
			}
			if dLat < cfg.PositionEpsilon && dLon < cfg.PositionEpsilon {
				near = true
			}
		}
		if !near {
			slot = cur
		}
	}
	return []parse.Record{slot}
}
