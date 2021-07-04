package main

// import  "github.com/guptarohit/asciigraph"
// "github.com/bsipos/thist"
// https://stackoverflow.com/questions/2471884/histogram-using-gnuplot
import (
	"math/rand"

	"github.com/Arafatk/glot"
	"github.com/andremueller/goreservoir/pkg/reservoir"
	"github.com/andremueller/goreservoir/pkg/sampling"
)

func increment() chan int {
	c := make(chan int)
	go func() {
		i := 0
		for {
			c <- i
			i++
		}
	}()
	return c
}

func indexArray(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

func indexArrayFloat64(n int) []float64 {
	result := make([]float64, n)
	for i := range result {
		result[i] = float64(i)
	}
	return result
}

func computeAges(i int, data []sampling.Sample) []int {
	ages := make([]int, len(data))
	for j := 0; j < len(data); j++ {
		ages[j] = i - data[j].(int)
	}
	return ages
}

func sum(data []int) int {
	s := 0
	for _, d := range data {
		s += d
	}
	return s
}

func scale(data []int) []float64 {
	scaled := make([]float64, len(data))
	max := sum(data)
	for i := range data {
		scaled[i] = float64(data[i]) / float64(max)
	}
	return scaled
}

func main() {
	rand.Seed(173)
	inc := increment()
	opts := reservoir.DynamicSamplerOpts{
		Lambda:   1.0 / 200.0,
		Capacity: 100,
	}
	sampler := sampling.NewChainSampler()
	sampler.AddLayer(reservoir.NewDynamic(opts))
	sampler.AddLayer(reservoir.NewDynamic(opts))
	maxIter := 20000
	allAges1 := make([]int, 2000)
	allAges2 := make([]int, 2000)
	for i := 0; i < maxIter; i++ {
		sampler.Add([]sampling.Sample{<-inc})
		if i >= 1000 && len(sampler.Layer(0).Data()) >= opts.Capacity-10 {
			updateAges(allAges1, computeAges(i, sampler.Layer(0).Data()))
		}
		if i >= 1000 && len(sampler.Layer(1).Data()) >= opts.Capacity-10 {
			updateAges(allAges2, computeAges(i, sampler.Layer(1).Data()))
		}
	}

	plot, err := glot.NewPlot(2, true, false)
	if err != nil {
		panic(err)
	}
	defer plot.Close()
	plot.AddPointGroup("ages_1", "lines", [][]float64{indexArrayFloat64(len(allAges1)), scale(allAges1)})
	plot.AddPointGroup("ages_2", "lines", [][]float64{indexArrayFloat64(len(allAges2)), scale(allAges2)})
}

func updateAges(summary []int, ages []int) {
	for _, a := range ages {
		if a >= 0 && a < len(summary) {
			summary[a]++
		}
	}
}
