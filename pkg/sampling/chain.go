package sampling

// ChainSampler is a chain of reservoirs whereas the output (discarded or not accepted points)
// from one stage is pasted to the next stage.
type ChainSampler struct {
	Layers []Sampler
}

func NewChainSampler() *ChainSampler {
	return &ChainSampler{
		Layers: make([]Sampler, 0, 5),
	}
}

func (s *ChainSampler) AddLayer(sampler Sampler) {
	s.Layers = append(s.Layers, sampler)
}

func (s *ChainSampler) Add(samples []Sample) []Sample {
	dat := samples
	for _, layer := range s.Layers {
		dat = layer.Add(dat)
	}
	return dat
}

func (s *ChainSampler) Data() []Sample {
	if len(s.Layers) == 0 {
		panic("empty layers")
	}
	return s.Layers[len(s.Layers)-1].Data()
}

var _ Sampler = (*ChainSampler)(nil)
