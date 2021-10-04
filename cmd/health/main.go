package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/andremueller/goreservoir/pkg/io"
	"github.com/andremueller/goreservoir/pkg/stats"
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

func unstripFloat(dat sampling.Sample) []float64 {
	data := dat.([]sampling.Sample)
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
	header     string
	outputFile string
	field      string
	lambda1    float64
	capacity1  int
	lambda2    float64
	capacity2  int
	minLen     int
	ensemble   int
	seed       int64
	skip       int
}

func (c *config) parse() {
	flag.StringVar(&c.inputFile, "input", "", "input csv file")
	flag.StringVar(&c.header, "header", "", "header to find (default: none)")
	flag.StringVar(&c.outputFile, "output", "result.csv", "output csv file")
	flag.StringVar(&c.field, "field", "", "input field to be used")
	flag.Float64Var(&c.lambda1, "lambda1", 1.0/1000.0, "Lambda of the first reservoir")
	flag.IntVar(&c.capacity1, "capacity1", 200, "Capacity of the first reservoir")
	flag.Float64Var(&c.lambda2, "lambda2", 1.0/10000.0, "Lambda of the second reservoir")
	flag.IntVar(&c.capacity2, "capacity2", 1000, "Capacity of the second reservoir")
	flag.IntVar(&c.minLen, "min", 50, "minimum capacity before running the metric")
	flag.IntVar(&c.ensemble, "ensemble", 1, "number of ensembles")
	flag.Int64Var(&c.seed, "seed", 0, "random seed value (0 = current time)")
	flag.IntVar(&c.skip, "skip", 1000, "Number of samples to wait before computing statistics")
	flag.Parse()
}

func createSampler(cfg *config) *sampling.EnsembleSampler {
	opts1 := reservoir.DynamicSamplerOpts{
		Lambda:   cfg.lambda1,
		Capacity: cfg.capacity1,
	}
	opts2 := reservoir.DynamicSamplerOpts{
		Lambda:   cfg.lambda2,
		Capacity: cfg.capacity2,
	}
	sampler := sampling.NewEnsembleSampler()

	for i := 0; i < cfg.ensemble; i++ {
		s := sampling.NewChainSampler()
		s.AddLayer(reservoir.NewDynamic(opts1))
		s.AddLayer(reservoir.NewDynamic(opts2))
		sampler.AddSampler(s)
	}

	return sampler
}

func computeAllStats(data []sampling.Sample, minLen int) []float64 {
	result := make([]float64, 0, len(data))
	for _, s := range data {
		samp := s.([]sampling.Sample)
		if len(samp) != 2 {
			panic("Wrong dimensionality")
		}
		v0 := unstripFloat(samp[0])
		v1 := unstripFloat(samp[1])

		if len(v0) >= minLen && len(v1) >= minLen {
			result = append(result, computeStat(v0, v1))
		}
	}
	return result
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	var cfg config
	cfg.parse()

	zap.S().Infof("Config: %+v", cfg)

	data, err := loadData(cfg.inputFile, cfg.header)
	if err != nil {
		panic(err)
	}
	zap.S().Infof("Loaded %d records (%d columns) from %s", data.Nrow(), data.Ncol(), cfg.inputFile)

	t := time.Now().Unix()
	if cfg.seed != 0 {
		t = cfg.seed
	}
	zap.S().Infof("Using random seed %d", t)
	rand.Seed(t)

	sampler := createSampler(&cfg)

	sel := data.Select([]string{cfg.field})

	indexValues := series.New(nil, series.Int, "index")
	statMean := series.New(nil, series.Float, "ks.stat")
	statMin := series.New(nil, series.Float, "ks.stat.min")
	statMax := series.New(nil, series.Float, "ks.stat.max")
	statStdDev := series.New(nil, series.Float, "ks.stat.sd")

	bar := progressbar.New(sel.Nrow())
	for i := 0; i < sel.Nrow(); i++ {
		bar.Add(1)
		sampler.Add([]sampling.Sample{sel.Elem(i, 0).Float()})
		if i >= cfg.skip {
			value := computeAllStats(sampler.Data(), cfg.minLen)
			fivenum := stats.ComputeFiveNum(value)
			indexValues.Append(i)
			statMean.Append(fivenum.Mean)
			statMin.Append(fivenum.Min)
			statMax.Append(fivenum.Max)
			statStdDev.Append(fivenum.StdDev)
		}
	}

	result := dataframe.New(indexValues, statMean, statMin, statMax, statStdDev)

	zap.S().Infof("Writing %d rows to output file %s", result.Nrow(), cfg.outputFile)
	writer, err := os.OpenFile(cfg.outputFile, os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	writer.WriteString(fmt.Sprintf("# config: %+v\n", cfg))
	result.WriteCSV(writer)
}
