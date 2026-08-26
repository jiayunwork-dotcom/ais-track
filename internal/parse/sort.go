package parse

import "sort"

func SortByTime(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Timestamp.Before(recs[j].Timestamp)
	})
}

func SortByVesselThenTime(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].MMSI != recs[j].MMSI {
			return recs[i].MMSI < recs[j].MMSI
		}
		return recs[i].Timestamp.Before(recs[j].Timestamp)
	})
}

func SortBySOG(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].SOG > recs[j].SOG
	})
}

func TopN(recs []Record, n int) []Record {
	if n >= len(recs) {
		return recs
	}
	return recs[:n]
}

func UniqueVessels(recs []Record) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, r := range recs {
		if _, ok := seen[r.MMSI]; !ok {
			seen[r.MMSI] = struct{}{}
			result = append(result, r.MMSI)
		}
	}
	return result
}

func CountByVessel(recs []Record) map[string]int {
	counts := make(map[string]int)
	for _, r := range recs {
		counts[r.MMSI]++
	}
	return counts
}
