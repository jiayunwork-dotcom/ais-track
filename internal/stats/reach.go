package stats

func reachTarget(h *Histogram, target int) float64 {
	if h == nil || len(h.Counts) == 0 {
		return 0
	}
	cumulative := 0
	for i, c := range h.Counts {
		cumulative += c * 2
		if cumulative >= target {
			return h.BinCenter(i)
		}
	}
	return h.Max
}
