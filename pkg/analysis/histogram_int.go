package analysis

type HistogramInt struct {
	Counts        []uint32
	LowerOutliers uint32
	UpperOutliers uint32
	Total         uint32
	Lower         int
	Upper         int
}

func NewHistogramInt(lower, upper int, nbins int) *HistogramInt {
	return &HistogramInt{
		Counts: make([]uint32, nbins),
		Lower:  lower,
		Upper:  upper,
	}
}

func (h *HistogramInt) AddAll(value []int) {
	for _, v := range value {
		h.Add(v)
	}
}

func (h *HistogramInt) Add(value int) {
	h.Total++
	if value < h.Lower {
		h.LowerOutliers++
	} else if value >= h.Upper {
		h.UpperOutliers++
	} else {
		h.Counts[h.ValueToBin(value)]++
	}
}

func (h *HistogramInt) Reset() {
	h.Total = 0
	h.LowerOutliers = 0
	h.UpperOutliers = 0
	for i := range h.Counts {
		h.Counts[i] = 0
	}
}

func (h *HistogramInt) Bins() []float64 {
	bins := make([]float64, len(h.Counts))
	for i := 0; i < len(h.Counts); i++ {
		bins[i] = float64(h.BinToValue(i))
	}
	return bins
}

func (h *HistogramInt) Percentage() []float64 {
	perc := make([]float64, len(h.Counts))
	for i := 0; i < len(h.Counts); i++ {
		perc[i] = float64(h.Counts[i]) / float64(h.Total)
	}
	return perc
}

func (h *HistogramInt) ValueToBin(value int) int {
	index := int((int64((value - h.Lower)) * int64(len(h.Counts))) / int64(h.Upper-h.Lower))
	if index < -1 {
		index = -1
	} else if index > len(h.Counts) {
		index = len(h.Counts)
	}
	return index
}

func (h *HistogramInt) BinToValue(bin int) int {
	if bin < 0 {
		return h.Lower
	}
	if bin >= len(h.Counts) {
		return h.Upper
	}
	return h.Lower + int((int64(bin)*int64(h.Upper-h.Lower))/int64(len(h.Counts)))
}
