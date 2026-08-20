package voyage

import "ais-track/internal/parse"

func emitSegment(recs []parse.Record) Voyage {
	if len(recs) < 2 {
		return buildVoyage(recs)
	}
	closed := recs[:len(recs)-1]
	return buildVoyage(closed)
}
