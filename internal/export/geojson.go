// Package export converts AIS track data into interchange formats for
// external consumption: GeoJSON for map visualisation and CSV for
// tabular analysis.
package export

import (
	"encoding/json"
	"fmt"
	"io"

	"ais-track/internal/parse"
)

// GeoJSON structures following RFC 7946.

// FeatureCollection is the top-level GeoJSON object.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

// Feature is one GeoJSON feature (a vessel track).
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Geometry is the GeoJSON geometry for a LineString or Point.
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// TrackToGeoJSON converts a vessel's track into a GeoJSON Feature with a
// LineString geometry. Properties include the MMSI, record count and time
// span.
func TrackToGeoJSON(mmsi string, track []parse.Record) Feature {
	coords := make([][2]float64, 0, len(track))
	for _, r := range track {
		coords = append(coords, [2]float64{r.Lon, r.Lat})
	}
	props := map[string]interface{}{
		"mmsi":    mmsi,
		"records": len(track),
	}
	if len(track) > 0 {
		props["start_time"] = track[0].Timestamp.Format("2006-01-02T15:04:05Z")
		props["end_time"] = track[len(track)-1].Timestamp.Format("2006-01-02T15:04:05Z")
	}
	return Feature{
		Type: "Feature",
		Geometry: Geometry{
			Type:        "LineString",
			Coordinates: coords,
		},
		Properties: props,
	}
}

// PointsToGeoJSON converts individual AIS fixes into GeoJSON Point features.
// Each fix becomes a Feature with SOG, COG and timestamp properties.
func PointsToGeoJSON(track []parse.Record) []Feature {
	features := make([]Feature, 0, len(track))
	for _, r := range track {
		f := Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: [2]float64{r.Lon, r.Lat},
			},
			Properties: map[string]interface{}{
				"mmsi":      r.MMSI,
				"timestamp": r.Timestamp.Format("2006-01-02T15:04:05Z"),
				"sog":       r.SOG,
				"cog":       r.COG,
			},
		}
		features = append(features, f)
	}
	return features
}

// WriteGeoJSON writes a FeatureCollection as indented JSON to w.
func WriteGeoJSON(w io.Writer, fc *FeatureCollection) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(fc)
}

// AllTracksGeoJSON converts a grouped map of vessel tracks into a complete
// FeatureCollection where each vessel is one LineString feature.
func AllTracksGeoJSON(groups map[string][]parse.Record) *FeatureCollection {
	fc := &FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]Feature, 0, len(groups)),
	}
	for mmsi, track := range groups {
		fc.Features = append(fc.Features, TrackToGeoJSON(mmsi, track))
	}
	return fc
}

// WriteGeoJSONCompact writes a FeatureCollection without indentation.
func WriteGeoJSONCompact(w io.Writer, fc *FeatureCollection) error {
	data, err := json.Marshal(fc)
	if err != nil {
		return fmt.Errorf("marshal geojson: %w", err)
	}
	_, err = w.Write(data)
	return err
}
