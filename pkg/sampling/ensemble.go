package sampling

// EnsembleSampler feeds a sample into multiple samplers
type EnsembleSampler struct {
	samplers []Sampler
}

func NewEnsembleSampler() *EnsembleSampler {
	return &EnsembleSampler{
		samplers: make([]Sampler, 0),
	}
}

// AddLayer adds a sampling layer to the end of the chain.
func (s *EnsembleSampler) AddSampler(sampler Sampler) {
	s.samplers = append(s.samplers, sampler)
}

func (s *EnsembleSampler) Sampler(i int) Sampler {
	return s.samplers[i]
}

func (s *EnsembleSampler) Count() int {
	return len(s.samplers)
}

func (s *EnsembleSampler) Add(samples []Sample) []Sample {
	dropped := make([]Sample, 0)
	for _, sampler := range s.samplers {
		dropped = append(dropped, sampler.Add(samples)...)
	}
	return dropped
}

// Data returns a slice of the current samples within the Sampler. In the ensemble
// sampler this will be a nested array of samples.
func (s *EnsembleSampler) Data() []Sample {
	if len(s.samplers) == 0 {
		panic("empty layers")
	}
	result := make([]Sample, 0)
	for _, sampler := range s.samplers {
		result = append(result, sampler.Data())
	}
	return result
}

var _ Sampler = (*EnsembleSampler)(nil)
