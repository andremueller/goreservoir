package main

// import  "github.com/guptarohit/asciigraph"
// "github.com/bsipos/thist"
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

func computeAges(i int, data []sampling.Sample) []int {
	ages := make([]int, len(data))
	for j := 0; j < len(data); j++ {
		ages[j] = i - data[j].(int)
	}
	return ages
}

func main() {
	rand.Seed(173)
	inc := increment()
	opts := reservoir.DynamicSamplerOpts{
		Lambda:   1.0 / 200.0,
		Capacity: 50,
	}
	sampler := reservoir.NewDynamic(opts)
	maxIter := 5000
	plot, err := glot.NewPlot(1, false, false)
	if err != nil {
		panic(err)
	}
	allAges := make([]int, 0, 5000)
	for i := 0; i < maxIter; i++ {
		sampler.Add([]sampling.Sample{<-inc})
		if i >= 1000 && len(sampler.Data()) >= opts.Capacity-10 {
			ages := computeAges(i, sampler.Data())
			allAges = append(allAges, ages...)
		}
	}
	plot.AddPointGroup("ages", "histogram", allAges)
	plot.SavePlot("histogram.png")
}
