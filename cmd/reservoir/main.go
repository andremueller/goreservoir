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

func main() {
	rand.Seed(173)
	nlayer := 5
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
		hists[i] = analysis.NewHistogramInt(0, 8000, 1000)
	}
	maxIter := 50000
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
	plot.SetTitle(fmt.Sprintf("Reservoir Sampling Chain \\lambda=%.3e, \\s=%d n=%d",
		opts.Lambda, opts.Capacity, maxIter))
	plot.SetYLabel("relative frequency")
	plot.SetXLabel("age [iterations]")
	defer plot.Close()
	for i, h := range hists {
		plot.AddPointGroup(fmt.Sprintf("ages_%d", i), "lines", [][]float64{h.Bins(), h.Percentage()})
	}
	err = plot.SavePlot("out.jpeg")
	if err != nil {
		panic(err)
	}
}

// computeAges returns a list of ages (in number of iterations) of the points in the reservoir.
// iter is the current iteration 0, 1, ...
func computeAges(iter int, data []sampling.Sample) []int {
	ages := make([]int, len(data))
	for j := 0; j < len(data); j++ {
		ages[j] = iter - data[j].(int)
	}
	return ages
}
