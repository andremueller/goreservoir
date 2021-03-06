package reservoir

import (
	"math"
	"math/rand"

	"github.com/andremueller/goreservoir/pkg/sampling"
)

// The smallest allowed lambda. The theoretically reservoir size is
const MinLambda = 1e-9

// DynamicSamplerOpts are the options for the DynamicSampler instance.
type DynamicSamplerOpts struct {
	// Lambda is the true lambda i.e. the forgetting factor greater than 0 and smaller than 1.
	Lambda float64

	// Capacity is the maximum number of entries in the reservoir. It can be smaller than the maximum theoretical
	// reservoir size of 1 / Lambda.
	Capacity int
}

// DynamicSampler is a biased variable reservoir sampler according to Algorithm 3.1 in
//
// Aggarwal, C. C. On biased reservoir sampling in the presence of stream evolution. in Proceedings of the 32nd international conference on Very large data bases 607–618 (ACM Press, 2006).
type DynamicSampler struct {
	opts      DynamicSamplerOpts // algorithm options
	reservoir []sampling.Sample  // the reservoir samples
	pMin      float64            // pMin is the lower bound of the acceptance probability
	pIn       float64            // pIn is the current acceptance probability
	qq        float64            // qq is the fraction of data points to be dropped
}

// Creates a new DynamicSampler instance with the given options `opts`.
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
		pIn:       1.0,
		qq:        1.0 / float64(opts.Capacity),
	}
}

func (s *DynamicSampler) Add(samples []sampling.Sample) []sampling.Sample {
	dropped := make([]sampling.Sample, 0)
	for _, sample := range samples {
		d := s.addSingle(sample)
		if d != nil {
			dropped = append(dropped, d...)
		}
	}
	return dropped
}

func (s *DynamicSampler) Data() []sampling.Sample {
	return s.reservoir
}

func (s *DynamicSampler) Reset() {
	s.reservoir = s.reservoir[:0]
}

// Adds a single sample to the reservoir.
// Returns a list of dropped samples containing those rejected by the sampling method before adding those to
// the reservoir and old samples which are removed from the reservoir.
func (s *DynamicSampler) addSingle(sample sampling.Sample) []sampling.Sample {
	dropped := make([]sampling.Sample, 0)
	nCur := len(s.reservoir)
	nVirt := math.Round(s.pIn / s.opts.Lambda) // virtual reservoir size (can be larger than the capacity)
	fillProp := float64(nCur) / float64(nVirt) // fill propability
	if dice(s.pIn) {
		// accept point
		if dice(fillProp) {
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

// removes the i-th sample from the array and returns the removed sample.
func remove(slice []sampling.Sample, i int) []sampling.Sample {
	// replace the element to be removed with the last element and shrink the slice by 1
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// dice returns true with probability `prob`.
func dice(prob float64) bool {
	return rand.Float64() <= prob
}

// compile time check for checking if DynamicSampler implements sampling.Sampler
// see https://go.dev/doc/faq#implements_interface
var _ sampling.Sampler = (*DynamicSampler)(nil)
