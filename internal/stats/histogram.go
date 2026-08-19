package stats

import (
	"fmt"
	"io"
	"math"
)

// Histogram bins a slice of float64 values into equal-width buckets for
// speed distribution analysis.
type Histogram struct {
	Min      float64
	Max      float64
	BinWidth float64
	Counts   []int
	Total    int
}

// NewHistogram creates a histogram with the given number of bins spanning
// [min, max]. Values outside this range are clipped to the edge bins.
func NewHistogram(min, max float64, bins int) *Histogram {
	if bins < 1 {
		bins = 10
	}
	if max <= min {
		max = min + 1
	}
	return &Histogram{
		Min:      min,
		Max:      max,
		BinWidth: (max - min) / float64(bins),
		Counts:   make([]int, bins),
	}
}

// Add inserts a value into the appropriate bin.
func (h *Histogram) Add(v float64) {
	h.Total++
	idx := int((v - h.Min) / h.BinWidth)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(h.Counts) {
		idx = len(h.Counts) - 1
	}
	h.Counts[idx]++
}

// AddAll inserts multiple values.
func (h *Histogram) AddAll(vals []float64) {
	for _, v := range vals {
		h.Add(v)
	}
}

// BinEdge returns the lower edge of the i-th bin.
func (h *Histogram) BinEdge(i int) float64 {
	return h.Min + float64(i)*h.BinWidth
}

// BinCenter returns the center of the i-th bin.
func (h *Histogram) BinCenter(i int) float64 {
	return h.BinEdge(i) + h.BinWidth/2
}

// MaxCount returns the maximum count in any bin.
func (h *Histogram) MaxCount() int {
	max := 0
	for _, c := range h.Counts {
		if c > max {
			max = c
		}
	}
	return max
}

// Percentile returns the approximate p-th percentile (0 < p < 1) from
// the histogram using linear interpolation.
func (h *Histogram) Percentile(p float64) float64 {
	if h.Total == 0 || p <= 0 || p >= 1 {
		return 0
	}
	target := int(math.Ceil(float64(h.Total) * p))
	cumulative := 0
	for i, c := range h.Counts {
		cumulative += c
		if cumulative >= target {
			return h.BinCenter(i)
		}
	}
	return h.Max
}

// WriteTo writes a text-based representation of the histogram to w.
func (h *Histogram) WriteTo(w io.Writer) (int64, error) {
	var total int64
	maxC := h.MaxCount()
	barWidth := 40
	for i, c := range h.Counts {
		lo := h.BinEdge(i)
		hi := lo + h.BinWidth
		bar := ""
		if maxC > 0 {
			n := c * barWidth / maxC
			for j := 0; j < n; j++ {
				bar += "#"
			}
		}
		line := fmt.Sprintf("[%6.1f, %6.1f) %5d %s\n", lo, hi, c, bar)
		n, err := io.WriteString(w, line)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
