package sampling

// ChainSampler is a chain of reservoirs whereas the output (discarded or not accepted points)
// from one stage is pasted to the next stage.
type ChainSampler struct {
	layers []Sampler
}

func NewChainSampler() *ChainSampler {
	return &ChainSampler{
		layers: make([]Sampler, 0),
	}
}

// AddLayer adds a sampling layer to the end of the chain.
func (s *ChainSampler) AddLayer(sampler Sampler) {
	s.layers = append(s.layers, sampler)
}

func (s *ChainSampler) Layer(i int) Sampler {
	return s.layers[i]
}

func (s *ChainSampler) Count() int {
	return len(s.layers)
}

func (s *ChainSampler) Add(samples []Sample) []Sample {
	dat := samples
	for _, layer := range s.layers {
		if dat == nil {
			return dat
		}
		dat = layer.Add(dat)
	}
	return dat
}

// Data returns a slice of the current samples within the Sampler. For a chain sampler
// this will be a list of all layers
func (s *ChainSampler) Data() []Sample {
	result := make([]Sample, len(s.layers))
	for i, layer := range s.layers {
		result[i] = layer.Data()
	}
	return result
}

func (s *ChainSampler) Reset() {
	for _, sampler := range s.layers {
		sampler.Reset()
	}
}

var _ Sampler = (*ChainSampler)(nil)
