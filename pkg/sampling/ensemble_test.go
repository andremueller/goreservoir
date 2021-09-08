package sampling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type recordingSampler struct {
	data []Sample
}

func newRecordingSampler() *recordingSampler {
	return &recordingSampler{
		data: make([]Sample, 0),
	}
}

func (s *recordingSampler) Add(samples []Sample) []Sample {
	s.data = append(s.data, samples...)
	return nil
}

func (s *recordingSampler) Data() []Sample {
	return s.data
}

func sampleToInt(value []Sample) []int {
	result := make([]int, len(value))
	for i, sample := range value {
		result[i] = sample.(int)
	}
	return result
}

func TestEnsembleSampler(t *testing.T) {
	ensemble := NewEnsembleSampler()
	ensemble.AddSampler(newRecordingSampler())
	ensemble.AddSampler(newRecordingSampler())

	expected := make([]int, 0)
	for i := 0; i < 10; i++ {
		ensemble.Add([]Sample{i})
		expected = append(expected, i)
	}
	result := ensemble.Data()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, expected, sampleToInt(result[0].([]Sample)))
	assert.Equal(t, expected, sampleToInt(result[1].([]Sample)))
}
