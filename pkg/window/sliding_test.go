package window

import (
	"testing"

	"github.com/andremueller/goreservoir/pkg/sampling"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	data := []sampling.Sample{1, 2, 3, 4, 5, 6, 7}
	s := NewSliding(5)
	assert.Equal(t, 5, s.Capacity())
	assert.Empty(t, s.Data())
	dropped := s.Add(data)
	assert.Equal(t, 5, len(s.Data()))
	dropped_expected := []sampling.Sample{1, 2}
	data_expected := []sampling.Sample{3, 4, 5, 6, 7}

	assert.Equal(t, dropped_expected, dropped)
	assert.Equal(t, data_expected, s.Data())
}
