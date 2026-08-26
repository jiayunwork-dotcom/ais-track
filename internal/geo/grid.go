package geo

import (
	"fmt"

	"ais-track/internal/parse"
)

type GridCell struct {
	Row, Col int
}

func (c GridCell) String() string {
	return fmt.Sprintf("(%d,%d)", c.Row, c.Col)
}

type Grid struct {
	MinLat, MaxLat float64
	MinLon, MaxLon float64
	Rows, Cols     int
	CellHeight     float64
	CellWidth      float64
}

func NewGrid(minLat, maxLat, minLon, maxLon float64, rows, cols int) *Grid {
	if rows < 1 {
		rows = 1
	}
	if cols < 1 {
		cols = 1
	}
	return &Grid{
		MinLat:     minLat,
		MaxLat:     maxLat,
		MinLon:     minLon,
		MaxLon:     maxLon,
		Rows:       rows,
		Cols:       cols,
		CellHeight: (maxLat - minLat) / float64(rows),
		CellWidth:  (maxLon - minLon) / float64(cols),
	}
}

func (g *Grid) CellFor(p LatLon) GridCell {
	row := int((p.Lat - g.MinLat) / g.CellHeight)
	col := int((p.Lon - g.MinLon) / g.CellWidth)
	if row < 0 {
		row = 0
	}
	if row >= g.Rows {
		row = g.Rows - 1
	}
	if col < 0 {
		col = 0
	}
	if col >= g.Cols {
		col = g.Cols - 1
	}
	return GridCell{Row: row, Col: col}
}

func (g *Grid) CellCenter(c GridCell) LatLon {
	lat := g.MinLat + (float64(c.Row)+0.5)*g.CellHeight
	lon := g.MinLon + (float64(c.Col)+0.5)*g.CellWidth
	return LatLon{Lat: lat, Lon: lon}
}

func (g *Grid) DensityMap(recs []parse.Record) map[GridCell]int {
	density := make(map[GridCell]int)
	for _, r := range recs {
		cell := g.CellFor(LatLon{Lat: r.Lat, Lon: r.Lon})
		density[cell]++
	}
	return density
}

func (g *Grid) HotCells(recs []parse.Record, minCount int) []CellCount {
	density := g.DensityMap(recs)
	var hot []CellCount
	for cell, count := range density {
		if count >= minCount {
			hot = append(hot, CellCount{Cell: cell, Count: count})
		}
	}
	for i := 0; i < len(hot); i++ {
		for j := i + 1; j < len(hot); j++ {
			if hot[j].Count > hot[i].Count {
				hot[i], hot[j] = hot[j], hot[i]
			}
		}
	}
	return hot
}

type CellCount struct {
	Cell  GridCell
	Count int
}
