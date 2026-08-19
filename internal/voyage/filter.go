package voyage

import "time"

// Filter defines criteria for selecting voyages from a set. All non-zero
// fields must be satisfied (logical AND).
type Filter struct {
	MinDuration  time.Duration
	MaxDuration  time.Duration
	MinDistance   float64
	MaxDistance   float64
	MinRecords   int
	MinMaxSOG    float64
}

// Match returns true if the voyage satisfies all non-zero filter criteria.
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

// FilterVoyages returns voyages that satisfy the filter. It does not
// modify the input slice.
func FilterVoyages(voyages []Voyage, f *Filter) []Voyage {
	if f == nil {
		result := make([]Voyage, len(voyages))
		copy(result, voyages)
		return result
	}
	var out []Voyage
	for i := range voyages {
		if f.Match(&voyages[i]) {
			out = append(out, voyages[i])
		}
	}
	return out
}

// ByDuration sorts voyages by duration descending (longest first).
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

// ByDistance sorts voyages by distance descending (longest first).
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
