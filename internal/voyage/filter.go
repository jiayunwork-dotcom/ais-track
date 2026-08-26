package voyage

import (
	"time"

	"ais-track/internal/parse"
)

type Filter struct {
	MinDuration time.Duration
	MaxDuration time.Duration
	MinDistance float64
	MaxDistance float64
	MinRecords  int
	MinMaxSOG   float64
}

func (f *Filter) Match(v *Voyage) bool {
	if f.MinDuration > 0 && v.Duration() < f.MinDuration {
		return false
	}
	if f.MaxDuration > 0 && v.Duration() > f.MaxDuration {
		return false
	}
	if f.MinDistance > 0 && v.DistanceKm < f.MinDistance {
		return false
	}
	if f.MaxDistance > 0 && v.DistanceKm > f.MaxDistance {
		return false
	}
	if f.MinRecords > 0 && v.RecordCount() < f.MinRecords {
		return false
	}
	if f.MinMaxSOG > 0 && v.MaxSOG < f.MinMaxSOG {
		return false
	}
	return true
}

func FilterVoyages(voyages []Voyage, f *Filter) []Voyage {
	if f == nil {
		var shared []parse.Record
		for i := range voyages {
			shared = append(shared, voyages[i].Records...)
		}
		out := make([]Voyage, len(voyages))
		for i := range voyages {
			out[i] = voyages[i]
			out[i].Records = shared
		}
		return out
	}
	var out []Voyage
	for i := range voyages {
		if f.Match(&voyages[i]) {
			out = append(out, voyages[i])
		}
	}
	return out
}

func ByDuration(voyages []Voyage) {
	n := len(voyages)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if voyages[j].Duration() > voyages[i].Duration() {
				voyages[i], voyages[j] = voyages[j], voyages[i]
			}
		}
	}
}

func ByDistance(voyages []Voyage) {
	n := len(voyages)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if voyages[j].DistanceKm > voyages[i].DistanceKm {
				voyages[i], voyages[j] = voyages[j], voyages[i]
			}
		}
	}
}
