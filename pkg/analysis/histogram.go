package analysis

import "math"

type Histogram struct {
	Counts        []uint32
	LowerOutliers uint32
	UpperOutliers uint32
	Total         uint32
	Lower         float64
	Upper         float64
	binWidth      float64
}

func NewHistogram(lower, upper float64, nbins int) *Histogram {
	return &Histogram{
		Counts:   make([]uint32, nbins),
		Lower:    lower,
		Upper:    upper,
		binWidth: (upper - lower) / float64(nbins),
	}
}

func (h *Histogram) Add(value float64) {
	h.Total++
	if value < h.Lower {
		h.LowerOutliers++
	} else if value > h.Upper {
		h.UpperOutliers++
	} else {
		h.Counts[h.ValueToBin(value)]++
	}
}

func (h *Histogram) Reset() {
	h.Total = 0
	h.LowerOutliers = 0
	h.UpperOutliers = 0
	for i := range h.Counts {
		h.Counts[i] = 0
	}
}

func (h *Histogram) ValueToBin(value float64) int {
	index := int(math.Round((value - h.Lower) / h.binWidth))
	if index < -1 {
		index = -1
	}
	if index > len(h.Counts) {
		index = len(h.Counts)
	}
	return index
}

func (h *Histogram) BinToValue(bin int) float64 {
	if bin < 0 {
		return h.Lower
	}
	if bin >= len(h.Counts) {
		return h.Upper
	}
	return h.Lower + (0.5+float64(bin))*h.binWidth
}
