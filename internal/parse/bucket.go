package parse

func putVessel(out map[string][]Record, r Record) {
	key := r.MMSI
	if key == "" {
		key = "_"
	}
	out[key] = append(out[key], r)
}
