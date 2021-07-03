package main

// import  "github.com/guptarohit/asciigraph"
import (
	"fmt"
	"math/rand"

	"github.com/andremueller/goreservoir/pkg/reservoir"
	"github.com/andremueller/goreservoir/pkg/sampling"
	"github.com/bsipos/thist"
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

func computeAges(i int, data []sampling.Sample) []float64 {
	ages := make([]float64, len(data))
	for j := 0; j < len(data); j++ {
		ages[j] = float64(i - data[j].(int))
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
	h := thist.NewHist(nil, "Age", "auto", -1, true)
	maxIter := 5000
	for i := 0; i < maxIter; i++ {
		sampler.Add([]sampling.Sample{<-inc})
		if i >= 1000 && len(sampler.Data()) >= opts.Capacity-10 {
			ages := computeAges(i, sampler.Data())
			for _, age := range ages {
				h.Update(age)
			}
		}
	}
	fmt.Println("Histogram:")
	fmt.Println(h.Draw())
}
