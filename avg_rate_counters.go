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

// Countable an interface that is used by AvgRateCounter to get the
// running count
//
// We have 2 implementation of this interface:
// 	- A Prometheus Metric object
//  - A Atomic counter
type Countable interface {
	Count() uint64
}

type AtomicCounter struct {
	count atomic.Uint64
}

func NewAtomicCounter() *AtomicCounter {
	return &AtomicCounter{}
}

func (a *AtomicCounter) Count() uint64 {
	v := a.count.Load()
	return v
}

func (a *AtomicCounter) Add(v uint64) {
	a.count.Add(v)
}

// PromCountable is a small wrapper to get a Prometheus Metric object tp implement the Countable inteface
// *Important* This  handles `Vec` metrics by summing for all labels, extracting for one specific label or for all labels is not yet supported.
type PromCountable struct {
	counter prometheus.Collector
}

func (p *PromCountable) current() uint64 {
	metricChan := make(chan prometheus.Metric, 16)
	go func() {
		p.counter.Collect(metricChan)
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
	//counter        prometheus.Collector
	counter        Countable
	samplingWindow time.Duration
	unit           string
	bucketCount    uint64
	totals         *ring.Ring[uint64]
	actualTotal    uint64
	actualCount    uint64
}

func MustNewAvgRateCounter(counter Countable, samplingWindow time.Duration, period time.Duration, unit string) *AvgRateCounter {
	a, err := NewAvgRateCounter(counter, samplingWindow, period, unit)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAvgRateCounter AvgRateCounter can be used to extract the average rate of a Countable object
// over a period of time. The computation is to accumulate the instant metric each <samplingWindow>
// and obtain an average over <period> duration. The <period> must be greater than
// <samplingWindow> and should be a multiple of it (enforced).
//
// If you for example want to log the rate of something each 30s and the rate is checked each 1s,
// your <period> should be set to 30s.
func NewAvgRateCounter(counter Countable, samplingWindow time.Duration, period time.Duration, unit string) (*AvgRateCounter, error) {
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

	rate := &AvgRateCounter{
		counter:        counter,
		samplingWindow: samplingWindow,
		unit:           unit,
		bucketCount:    bucketCount,
		totals:         ring.New[uint64](int(bucketCount)),
	}

	// FIXME: See `run` documentation about the FIXME
	rate.run()

	return rate, nil
}

func (c *AvgRateCounter) Total() uint64 {
	return c.actualTotal
}

func (c *AvgRateCounter) RateInt64() int64 {
	return int64(c.rate())
}

func (c *AvgRateCounter) RateFloat64() float64 {
	return c.rate()
}

func (c *AvgRateCounter) RateString() string {
	return strconv.FormatFloat(c.RateFloat64(), 'f', -1, 64)
}

func (c *AvgRateCounter) String() string {
	return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, timeUnitToString(c.samplingWindow), c.Total())
}

func (c *AvgRateCounter) rate() float64 {
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
func (c *AvgRateCounter) run() {
	go func() {
		ticker := time.NewTicker(c.samplingWindow)
		defer ticker.Stop()

		for {
			<-ticker.C

			c.actualCount++
			c.actualTotal = c.counter.Count()

			c.totals.Value = c.actualTotal
			c.totals = c.totals.Next()
		}
	}()
}
