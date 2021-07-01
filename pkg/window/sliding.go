package window

import "github.com/andremueller/goreservoir/pkg/sampling"

// Sliding is a simple sliding window approach.
type Sliding struct {
	n      int
	buffer []sampling.Sample
}

func NewSliding(n int) *Sliding {
	return &Sliding{
		n:      n,
		buffer: make([]sampling.Sample, 0, n),
	}
}

func (s *Sliding) Capacity() int {
	return s.n
}

func (s *Sliding) Add(samples []sampling.Sample) []sampling.Sample {
	dropped := make([]sampling.Sample, 0)

	for _, sample := range samples {
		if drop := s.addSingle(sample); drop != nil {
			dropped = append(dropped, drop)
		}
	}

	return dropped
}

func (s *Sliding) addSingle(sample sampling.Sample) sampling.Sample {
	if len(s.buffer) < s.n {
		s.buffer = append(s.buffer, sample)
		return nil
	} else {
		dropped := s.buffer[0]
		s.buffer = append(s.buffer[1:], sample)
		return dropped
	}
}

func (s *Sliding) Data() []sampling.Sample {
	return s.buffer
}

var _ sampling.Sampler = (*Sliding)(nil)
