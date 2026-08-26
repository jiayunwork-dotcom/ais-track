package export

import (
	"encoding/json"
	"fmt"
	"io"

	"ais-track/internal/parse"
)

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

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

func WriteGeoJSON(w io.Writer, fc *FeatureCollection) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(fc)
}

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

func WriteGeoJSONCompact(w io.Writer, fc *FeatureCollection) error {
	data, err := json.Marshal(fc)
	if err != nil {
		return fmt.Errorf("marshal geojson: %w", err)
	}
	_, err = w.Write(data)
	return err
}
