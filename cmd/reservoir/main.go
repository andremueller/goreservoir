package main

// import  "github.com/guptarohit/asciigraph"
// "github.com/bsipos/thist"
// https://stackoverflow.com/questions/2471884/histogram-using-gnuplot
import (
	"fmt"
	"math/rand"

	"github.com/Arafatk/glot"
	"github.com/andremueller/goreservoir/pkg/analysis"
	"github.com/andremueller/goreservoir/pkg/reservoir"
	"github.com/andremueller/goreservoir/pkg/sampling"
)

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

func main() {
	rand.Seed(173)
	nlayer := 3
	opts := reservoir.DynamicSamplerOpts{
		Lambda:   1.0 / 200.0,
		Capacity: 100,
	}
	sampler := sampling.NewChainSampler()
	for i := 0; i < nlayer; i++ {
		sampler.AddLayer(reservoir.NewDynamic(opts))
	}
	hists := make([]*analysis.HistogramInt, nlayer)
	for i := range hists {
		hists[i] = analysis.NewHistogramInt(0, 2000, 1000)
	}
	maxIter := 20000
	for t := 0; t < maxIter; t++ {
		sampler.Add([]sampling.Sample{t})
		for j := 0; j < sampler.Count(); j++ {
			data := sampler.Layer(j).Data()
			if len(data) > opts.Capacity-10 {
				ages := computeAges(t, data)
				hists[j].AddAll(ages)
			}
		}
	}

	plot, err := glot.NewPlot(2, true, false)
	if err != nil {
		panic(err)
	}
	defer plot.Close()
	for i, h := range hists {
		plot.AddPointGroup(fmt.Sprintf("ages_%d", i), "lines", [][]float64{h.Bins(), h.Percentage()})
	}
}
