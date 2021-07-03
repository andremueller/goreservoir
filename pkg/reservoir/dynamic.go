package reservoir

import (
	"math"
	"math/rand"

	"github.com/andremueller/goreservoir/pkg/sampling"
)

const MinLambda = 1e-9

type DynamicSamplerOpts struct {
	// Lambda is the true lambda i.e. the forgetting factor between 0 and 1.
	Lambda float64

	// Capacity is the maximum number of entries in the reservoir
	Capacity int
}

// DynamicSampler is a biased variable reservoir sampler akkording to Algorithm 3.1
// in Aggarwal.
//
type DynamicSampler struct {
	opts      DynamicSamplerOpts
	reservoir []sampling.Sample
	pMin      float64 // pMin is the lower bound of the acceptance probability
	pIn       float64 // pIn is the current acceptance probability
	qq        float64 // qq is the fraction of data points to be dropped
}

func NewDynamic(opts DynamicSamplerOpts) *DynamicSampler {
	if opts.Lambda < MinLambda || opts.Lambda > 1.0 {
		panic("Lambda is out of range")
	}
	n := int(math.Ceil(1 / opts.Lambda))
	if opts.Capacity > n {
		// reduce maximum capacity to the theoretical boundary
		opts.Capacity = n
	}

	return &DynamicSampler{
		opts:      opts,
		reservoir: make([]sampling.Sample, 0, opts.Capacity),
		pMin:      math.Min(1.0, float64(opts.Capacity)/float64(n)),
		qq:        1.0 / float64(opts.Capacity),
	}
}

func (s *DynamicSampler) Add(samples []sampling.Sample) []sampling.Sample {
	dropped := make([]sampling.Sample, 0)
	for _, sample := range samples {
		d := s.addSingle(sample)
		if d != nil {
			dropped = append(dropped, d)
		}
	}
	return dropped
}

func (s *DynamicSampler) addSingle(sample sampling.Sample) []sampling.Sample {
	dropped := make([]sampling.Sample, 0)
	nCur := len(s.reservoir)
	nVirt := math.Round(s.pIn / s.opts.Lambda)
	fill := float64(nCur) / float64(nVirt)
	if dice(s.pIn) {
		// accept point
		if dice(fill) {
			// replace old point
			i := rand.Intn(nCur)
			dropped = append(dropped, s.reservoir[i])
			s.reservoir[i] = sample
		} else {
			// just append new point to reservoir
			s.reservoir = append(s.reservoir, sample)
		}
	}
	nCur = len(s.reservoir)
	if nCur > s.opts.Capacity {
		// Reservoir is full => reduce acceptance probability pIn until pMin is reached
		// Reduce acceptance probability pIn
		pInNew := s.pIn * (1.0 - s.qq)
		if pInNew > s.pMin {
			s.pIn = math.Max(s.pMin, pInNew)
			// eject the fraction qq of the old points
			nEject := maxInt(1, int(math.Ceil(float64(nCur)*s.qq)))
			for i := 0; i < nEject; i++ {
				index := rand.Intn(len(s.reservoir))
				dropped = append(dropped, s.reservoir[index])
				s.reservoir = remove(s.reservoir, index)
			}
		}
	}

	return dropped
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func remove(slice []sampling.Sample, i int) []sampling.Sample {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// dice returns true with probability `prob`.
func dice(prob float64) bool {
	return rand.Float64() <= prob
}

// Data returns a slice of the current samples within the Sampler.
func (s *DynamicSampler) Data() []sampling.Sample {
	return s.reservoir
}
