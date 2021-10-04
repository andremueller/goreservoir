package stats

import (
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

type FiveNum struct {
	Mean   float64
	Min    float64
	Max    float64
	StdDev float64
	Count  int
}

var fiveNumDefault = &FiveNum{
	Mean:   math.NaN(),
	Min:    math.NaN(),
	Max:    math.NaN(),
	StdDev: math.NaN(),
	Count:  0,
}

func ComputeFiveNum(values []float64) *FiveNum {
	if len(values) > 0 {
		result := &FiveNum{}

		result.Count = len(values)
		result.Min = floats.Min(values)
		result.Max = floats.Max(values)
		result.Mean = stat.Mean(values, nil)
		if len(values) >= 2 {
			result.StdDev = stat.StdDev(values, nil)
		} else {
			result.StdDev = math.NaN()
		}
		return result
	}
	return fiveNumDefault
}
