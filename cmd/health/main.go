package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/andremueller/goreservoir/pkg/io"
	"go.uber.org/zap"
	"gonum.org/v1/gonum/stat"

	"golang.org/x/text/transform"

	"github.com/andremueller/goreservoir/pkg/reservoir"
	"github.com/andremueller/goreservoir/pkg/sampling"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/rotisserie/eris"
	"github.com/schollz/progressbar"
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

type config struct {
	inputFile  string
	skipHeader string
	outputFile string
	field      string
	lambda1    float64
	capacity1  int
	lambda2    float64
	capacity2  int
	minLen     int
}

func (c *config) parse() {
	flag.StringVar(&c.inputFile, "input", "", "input csv file")
	flag.StringVar(&c.skipHeader, "skip", "", "header to find (default: none)")
	flag.StringVar(&c.outputFile, "output", "result.csv", "output csv file")
	flag.StringVar(&c.field, "field", "", "input field to be used")
	flag.Float64Var(&c.lambda1, "lambda1", 1.0/1000.0, "Lambda of the first reservoir")
	flag.IntVar(&c.capacity1, "capacity1", 200, "Capacity of the first reservoir")
	flag.Float64Var(&c.lambda2, "lambda2", 1.0/10000.0, "Lambda of the second reservoir")
	flag.IntVar(&c.capacity2, "capacity2", 1000, "Capacity of the second reservoir")
	flag.IntVar(&c.minLen, "min", 50, "minimum capacity before running the metric")
	flag.Parse()
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	var cfg config
	cfg.parse()

	data, err := loadData(cfg.inputFile, cfg.skipHeader)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("Loaded %d records (%d columns) from %s", data.Nrow(), data.Ncol(), cfg.inputFile)

	rand.Seed(173)
	opts1 := reservoir.DynamicSamplerOpts{
		Lambda:   cfg.lambda1,
		Capacity: cfg.capacity1,
	}
	opts2 := reservoir.DynamicSamplerOpts{
		Lambda:   cfg.lambda2,
		Capacity: cfg.capacity2,
	}
	sampler := sampling.NewChainSampler()
	sampler.AddLayer(reservoir.NewDynamic(opts1))
	sampler.AddLayer(reservoir.NewDynamic(opts2))

	sel := data.Select([]string{cfg.field})

	stat := make([]float64, 0)
	index := make([]int, 0)
	started := false
	bar := progressbar.New(sel.Nrow())
	for i := 0; i < sel.Nrow(); i++ {
		bar.Add(1)
		sampler.Add([]sampling.Sample{sel.Elem(i, 0).Float()})
		v0 := unstripFloat(sampler.Layer(0).Data())
		v1 := unstripFloat(sampler.Layer(1).Data())
		if len(v0) >= cfg.minLen && len(v1) >= cfg.minLen {
			if !started {
				zap.S().Infof("Started metric at row %d", i)
				started = true
			}
			stat = append(stat, computeStat(v1, v0))
			index = append(index, i)
		}
	}
	indexValues := series.New(index, series.Int, "index")
	statValues := series.New(stat, series.Float, "ks.stat")

	result := dataframe.New(indexValues, statValues)

	zap.S().Infof("Writing %d rows to output file %s", result.Nrow(), cfg.outputFile)
	writer, err := os.OpenFile(cfg.outputFile, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	writer.WriteString(fmt.Sprintf("# config: %+v\n", cfg))
	result.WriteCSV(writer)
}
