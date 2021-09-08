package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/andremueller/goreservoir/pkg/io"
	"gonum.org/v1/gonum/stat"

	"golang.org/x/text/transform"

	"github.com/andremueller/goreservoir/pkg/reservoir"
	"github.com/andremueller/goreservoir/pkg/sampling"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/rotisserie/eris"
)

func loadData(fileName string, findHeader string) (dataframe.DataFrame, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return dataframe.DataFrame{}, eris.Wrap(err, "os.Open")
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if len(findHeader) > 0 {
		err = io.ReaderSkipUntil(reader, "[Messdaten]", 100)
		if err != nil {
			return dataframe.DataFrame{}, eris.Wrap(err, "io.ReaderSkipUntil")
		}
	}
	trf := transform.NewReader(reader, io.NewReplacingTransformer(",", "."))
	data := dataframe.ReadCSV(trf, dataframe.HasHeader(true), dataframe.WithDelimiter('\t'), dataframe.DefaultType(series.Float))
	return data, nil
}

func unstripFloat(data []sampling.Sample) []float64 {
	result := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i].(float64)
	}
	return result
}

func computeStat(old, new []float64) float64 {
	sort.Float64s(old)
	sort.Float64s(new)
	return stat.KolmogorovSmirnov(old, nil, new, nil)
}

func main() {
	var inputFile string
	var skipHeader string
	var outputFile string
	var field string
	flag.StringVar(&inputFile, "input", "", "input csv file")
	flag.StringVar(&skipHeader, "skip", "", "header to find (default: none)")
	flag.StringVar(&outputFile, "output", "result.csv", "output csv file")
	flag.StringVar(&field, "field", "", "input field to be used")
	flag.Parse()
	data, err := loadData(inputFile, skipHeader)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d records (%d columns)", data.Nrow(), data.Ncol())

	rand.Seed(173)
	nlayer := 2
	opts := reservoir.DynamicSamplerOpts{
		Lambda:   1.0 / 2000.0,
		Capacity: 200,
	}
	sampler := sampling.NewChainSampler()
	for i := 0; i < nlayer; i++ {
		sampler.AddLayer(reservoir.NewDynamic(opts))
	}

	sel := data.Select([]string{field})
	startAt := 100

	stat := make([]float64, 0)
	index := make([]int, 0)
	for i := 0; i < sel.Nrow(); i++ {
		sampler.Add([]sampling.Sample{sel.Elem(i, 0).Float()})
		if i >= startAt {
			v0 := unstripFloat(sampler.Layer(0).Data())
			v1 := unstripFloat(sampler.Layer(1).Data())
			stat = append(stat, computeStat(v1, v0))
			index = append(index, i)
		}
	}
	indexValues := series.New(index, series.Int, "index")
	statValues := series.New(stat, series.Float, "ks.stat")

	result := dataframe.New(indexValues, statValues)

	writer, err := os.OpenFile(outputFile, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	result.WriteCSV(writer)
}
