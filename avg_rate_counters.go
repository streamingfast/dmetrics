package dmetrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/streamingfast/dmetrics/ring"
	"go.uber.org/atomic"
)

// *Important* This  handles `Vec` metrics by summing for all labels, extracting for one specific label or for all labels is not yet supported.

//
type AvgRatePromCounter struct {
	*avgRate
	collector prometheus.Collector
}

func MustNewAvgRateFromPromCounter(promCollector prometheus.Collector, samplingWindow time.Duration, period time.Duration, unit string) *AvgRatePromCounter {
	a, err := NewAvgRateFromPromCounter(promCollector, samplingWindow, period, unit)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAvgRateFromPromCounter Extracts the average rate of a Prom Collector
// over a period of time. The rate is computed by accumulating the total
// count at the <samplingWindow> interval and averaging them our ove the number
// of <period> defined.
// Suppose there is a block count that increments as follows
//
//	0s to 1s -> 10 blocks
//	1s to 2s -> 3 blocks
//	2s to 3s -> 0 blocks
//	3s to 4s -> 7 blocks
//
// If your  samplingWindow = 1s and your period = 4s, the rate will be computed as
//	(10 + 3 + 0 + 7)/4 = 5 blocks/sec
//
// If your  samplingWindow = 1s and your period = 3s, the rate will be computed as
//	(10 + 3 + 0)/4 = 4.33 blocks/sec
// then when the "window moves" you would get
//	(3 + 0 + 7)/4 = 3.333 blocks/sec
//
func NewAvgRateFromPromCounter(promCollector prometheus.Collector, samplingWindow time.Duration, period time.Duration, unit string) (*AvgRatePromCounter, error) {
	a := &AvgRatePromCounter{collector: promCollector}
	avgRage, err := newAvgRateCounter(a.count, samplingWindow, period, unit)
	if err != nil {
		return nil, fmt.Errorf("new avg rate counter: %w", err)
	}
	a.avgRate = avgRage
	return a, nil
}

func (a *AvgRatePromCounter) count() uint64 {
	metricChan := make(chan prometheus.Metric, 16)
	go func() {
		a.collector.Collect(metricChan)
		close(metricChan)
	}()

	sum := 0.0
	for value := range metricChan {
		model := new(dto.Metric)
		err := value.Write(model)
		if err != nil {
			panic(err)
		}

		if model.Counter != nil && model.Counter.Value != nil {
			sum += *model.Counter.Value
		}
	}

	return uint64(sum)
}

type AvgRateCounter struct {
	*avgRate
	c *atomic.Uint64
}

func MustNewAvgRateCounter(samplingWindow time.Duration, period time.Duration, unit string) *AvgRateCounter {
	a, err := NewAvgRateCounter(samplingWindow, period, unit)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAvgRateCounter Tracks the average rate over a period of time. The rate is
// computed by accumulating the total count at the <samplingWindow> interval and
// averaging them our ove the number of <period> defined.
// Suppose there is a block count that increments as follows
//
//	0s to 1s -> 10 blocks
//	1s to 2s -> 3 blocks
//	2s to 3s -> 0 blocks
//	3s to 4s -> 7 blocks
//
// If your  samplingWindow = 1s and your period = 4s, the rate will be computed as
//	(10 + 3 + 0 + 7)/4 = 5 blocks/sec
//
// If your  samplingWindow = 1s and your period = 3s, the rate will be computed as
//	(10 + 3 + 0)/4 = 4.33 blocks/sec
// then when the "window moves" you would get
//	(3 + 0 + 7)/4 = 3.333 blocks/sec
func NewAvgRateCounter(samplingWindow time.Duration, period time.Duration, unit string) (*AvgRateCounter, error) {
	a := &AvgRateCounter{c: atomic.NewUint64(0)}
	avgRage, err := newAvgRateCounter(a.count, samplingWindow, period, unit)
	if err != nil {
		return nil, fmt.Errorf("new avg rate counter: %w", err)
	}
	a.avgRate = avgRage
	return a, nil
}

// Add tracks a number of events, to be used to compute the rage
func (a *AvgRateCounter) Add(v uint64) {
	a.c.Add(v)
}

func (a *AvgRateCounter) count() uint64 {
	return a.c.Load()
}

type CountableFunc = func() uint64

// NewAvgRateCounter AvgRateCounter can be used to extract the average rate of a Countable object
// over a period of time. The computation is to accumulate the instant metric each <samplingWindow>
// and obtain an average over <period> duration. The <period> must be greater than
// <samplingWindow> and should be a multiple of it (enforced).
//
// If you for example want to log the rate of something each 30s and the rate is checked each 1s,
// your <period> should be set to 30s.

type avgRate struct {
	//counter        prometheus.Collector
	counterFunc    CountableFunc
	samplingWindow time.Duration
	unit           string
	bucketCount    uint64
	totals         *ring.Ring[uint64]
	actualTotal    uint64
	actualCount    uint64
}

func newAvgRateCounter(counter CountableFunc, samplingWindow time.Duration, period time.Duration, unit string) (*avgRate, error) {

	if samplingWindow == 0 {
		return nil, fmt.Errorf("sampling window must be greater then 0")
	}

	if samplingWindow > period {
		return nil, fmt.Errorf("interval (%s) must be lower than averageTime (%s) but it's not", samplingWindow, period)
	}

	if period%samplingWindow != 0 {
		return nil, fmt.Errorf("averageTime (%s) must be divisible by samplingWindow (%s) without a remainder but it's not", period, samplingWindow)
	}

	bucketCount := (uint64(period / samplingWindow)) + 1

	rate := &avgRate{
		counterFunc:    counter,
		samplingWindow: samplingWindow,
		unit:           unit,
		bucketCount:    bucketCount,
		totals:         ring.New[uint64](int(bucketCount)),
	}

	// FIXME: See `run` documentation about the FIXME
	rate.run()

	return rate, nil
}

func (c *avgRate) Total() uint64      { return c.actualTotal }
func (c *avgRate) Rate() float64      { return c.rate() }
func (c *avgRate) RateString() string { return strconv.FormatFloat(c.Rate(), 'f', 3, 64) }
func (c *avgRate) String() string {
	return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, timeUnitToString(c.samplingWindow), c.Total())
}
func (c *avgRate) rate() float64 {
	skip := uint64(0)
	if c.actualCount < uint64(c.bucketCount) {
		// We do an extra minus one because we are interested about delta and there is always `c.bucketCount - 1` deltas
		skip = c.bucketCount - c.actualCount - 1
	}

	var sum uint64
	var deltaCount uint64
	var valueCount uint64
	var previousData *uint64

	c.totals.Do(func(total uint64) {
		if valueCount > skip && previousData != nil {
			sum += total - *previousData
			deltaCount++
		}

		previousData = &total
		valueCount++
	})

	return float64(sum) / float64(deltaCount)
}

// FIXME: Use finalizer trick (search online) to stop the goroutine when the counter goes out of scope
// for now, lifecycle is not handled a rate from counter lives forever
func (c *avgRate) run() {
	go func() {
		ticker := time.NewTicker(c.samplingWindow)
		defer ticker.Stop()

		for {
			<-ticker.C

			c.actualCount++
			c.actualTotal = c.counterFunc()

			c.totals.Value = c.actualTotal
			c.totals = c.totals.Next()
		}
	}()
}
