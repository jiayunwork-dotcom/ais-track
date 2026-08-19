package parse

import (
	"strings"
	"testing"
	"time"
)

func TestParseRecordsValid(t *testing.T) {
	const data = "mmsi,ts,lat,lon,sog,cog\n" +
		"123456789,2023-01-01T00:00:00Z,1.0,2.0,5.5,90.0\n" +
		"123456789,2023-01-01T00:01:00Z,1.1,2.1,6.0,95.0\n"
	recs, err := ParseRecords(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("want 2 records, got %d", len(recs))
	}
	if recs[0].MMSI != "123456789" {
		t.Errorf("MMSI = %q", recs[0].MMSI)
	}
	if recs[0].Lat != 1.0 || recs[0].Lon != 2.0 {
		t.Errorf("lat/lon = %v/%v", recs[0].Lat, recs[0].Lon)
	}
	if recs[0].SOG != 5.5 || recs[0].COG != 90.0 {
		t.Errorf("sog/cog = %v/%v", recs[0].SOG, recs[0].COG)
	}
	want := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	if !recs[0].Timestamp.Equal(want) {
		t.Errorf("timestamp = %v, want %v", recs[0].Timestamp, want)
	}
}

func TestParseRecordsBadNumber(t *testing.T) {
	const data = "mmsi,ts,lat,lon,sog,cog\n" +
		"123456789,2023-01-01T00:00:00Z,notanumber,2.0,5.5,90.0\n"
	if _, err := ParseRecords(strings.NewReader(data)); err == nil {
		t.Fatal("expected error for bad latitude, got nil")
	}
}

func TestParseRecordsBadTimestamp(t *testing.T) {
	const data = "mmsi,ts,lat,lon,sog,cog\n" +
		"123456789,not-a-time,1.0,2.0,5.5,90.0\n"
	if _, err := ParseRecords(strings.NewReader(data)); err == nil {
		t.Fatal("expected error for bad timestamp, got nil")
	}
}

func TestParseRecordsBadHeader(t *testing.T) {
	const data = "foo,bar,baz,qux,quux,corge\n1,2,3,4,5,6\n"
	if _, err := ParseRecords(strings.NewReader(data)); err == nil {
		t.Fatal("expected error for bad header, got nil")
	}
}

func TestGroupByVesselNil(t *testing.T) {
	got := GroupByVessel(nil)
	if got == nil {
		t.Fatal("GroupByVessel(nil) returned nil map")
	}
	if len(got) != 0 {
		t.Fatalf("want empty map, got %d entries", len(got))
	}
}

func TestGroupByVesselGroups(t *testing.T) {
	recs := []Record{
		{MMSI: "A", Timestamp: time.Now()},
		{MMSI: "B", Timestamp: time.Now()},
		{MMSI: "A", Timestamp: time.Now()},
	}
	got := GroupByVessel(recs)
	if len(got) != 2 {
		t.Fatalf("want 2 vessels, got %d", len(got))
	}
	if len(got["A"]) != 2 {
		t.Fatalf("vessel A: want 2 records, got %d", len(got["A"]))
	}
	if len(got["B"]) != 1 {
		t.Fatalf("vessel B: want 1 record, got %d", len(got["B"]))
	}
}
