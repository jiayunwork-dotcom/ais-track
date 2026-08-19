package parse

import "sort"

// SortByTime sorts records by timestamp in ascending order. This is
// required before feeding records into voyage segmentation or gap
// detection.
func SortByTime(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Timestamp.Before(recs[j].Timestamp)
	})
}

// SortByVesselThenTime sorts records first by MMSI (ascending) then by
// timestamp within each vessel. Useful for bulk processing of multi-vessel
// datasets.
func SortByVesselThenTime(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].MMSI != recs[j].MMSI {
			return recs[i].MMSI < recs[j].MMSI
		}
		return recs[i].Timestamp.Before(recs[j].Timestamp)
	})
}

// SortBySOG sorts records by speed over ground in descending order.
func SortBySOG(recs []Record) {
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].SOG > recs[j].SOG
	})
}

// TopN returns the first n records from the slice. If n exceeds the
// length, all records are returned.
func TopN(recs []Record, n int) []Record {
	if n >= len(recs) {
		return recs
	}
	return recs[:n]
}

// UniqueVessels returns the distinct MMSI values from the records.
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

// CountByVessel returns the record count per vessel.
func CountByVessel(recs []Record) map[string]int {
	counts := make(map[string]int)
	for _, r := range recs {
		counts[r.MMSI]++
	}
	return counts
}
