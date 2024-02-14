package dmetrics

import (
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type avgRatePromCollector struct {
	*avgRate
	collector          prometheus.Collector
	promMetricsToValue func(metrics []*dto.Metric) uint64
}

func (c *avgRatePromCollector) Stop() {
	stopJanitor(c)
}

func (a *avgRatePromCollector) count() uint64 {
	metricChan := make(chan prometheus.Metric, 16)
	go func() {
		a.collector.Collect(metricChan)
		close(metricChan)
	}()

	metrics := make([]*dto.Metric, 0, 16)
	for value := range metricChan {
		model := new(dto.Metric)
		err := value.Write(model)
		if err != nil {
			panic(err)
		}

		metrics = append(metrics, model)
	}

	return a.promMetricsToValue(metrics)
}

// *Important* This  handles `Vec` metrics by summing for all labels, extracting for one specific label or for all labels is not yet supported.

type AvgRatePromCounter struct {
	*avgRatePromCollector
}

// MustNewAvgRateFromPromCounter acts like [NewAvgRateFromPromCounter] but panics if an error occurs.
// Refers to [NewAvgRateFromPromCounter] for more information.
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
//
//	(10 + 3 + 0 + 7)/4 = 5 blocks/sec
//
// If your  samplingWindow = 1s and your period = 3s, the rate will be computed as
//
//	(10 + 3 + 0)/4 = 4.33 blocks/sec
//
// then when the "window moves" you would get
//
//	(3 + 0 + 7)/4 = 3.333 blocks/sec
func NewAvgRateFromPromCounter(promCollector prometheus.Collector, samplingWindow time.Duration, period time.Duration, unit string) (*AvgRatePromCounter, error) {
	a, err := newAvgRateFromPromCollector(promCollector, samplingWindow, period, unit, func(metrics []*dto.Metric) (sum uint64) {
		for _, m := range metrics {
			if m.Counter != nil && m.Counter.Value != nil {
				sum += uint64(*m.Counter.Value)
			}
		}
		return
	})
	if err != nil {
		return nil, err
	}

	return &AvgRatePromCounter{a}, nil
}

type AvgRatePromGauge struct {
	*avgRatePromCollector
}

// MustNewAvgRateFromPromGauge acts like [NewAvgRateFromPromGauge] but panics if an error occurs.
// Refers to [NewAvgRateFromPromGauge] for more information.
func MustNewAvgRateFromPromGauge(promCollector prometheus.Collector, samplingWindow time.Duration, period time.Duration, unit string) *AvgRatePromGauge {
	a, err := NewAvgRateFromPromGauge(promCollector, samplingWindow, period, unit)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAvgRateFromPromGauge extracts the average rate of a Promtheus Gauge
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
//
//	(10 + 3 + 0 + 7)/4 = 5 blocks/sec
//
// If your  samplingWindow = 1s and your period = 3s, the rate will be computed as
//
//	(10 + 3 + 0)/4 = 4.33 blocks/sec
//
// then when the "window moves" you would get
//
//	(3 + 0 + 7)/4 = 3.333 blocks/sec
//
// **Important** Your Gauge should be ever increasing. If it's not, the results will be incorrect.
func NewAvgRateFromPromGauge(promCollector prometheus.Collector, samplingWindow time.Duration, period time.Duration, unit string) (*AvgRatePromGauge, error) {
	a, err := newAvgRateFromPromCollector(promCollector, samplingWindow, period, unit, func(metrics []*dto.Metric) (sum uint64) {
		for _, m := range metrics {
			if m.Gauge != nil && m.Gauge.Value != nil {
				sum += uint64(*m.Gauge.Value)
			}
		}
		return
	})
	if err != nil {
		return nil, err
	}

	return &AvgRatePromGauge{a}, nil
}

func newAvgRateFromPromCollector(
	promCollector prometheus.Collector,
	samplingWindow time.Duration,
	period time.Duration,
	unit string,
	promMetricsToValue func(metrics []*dto.Metric) uint64,
) (*avgRatePromCollector, error) {
	a := &avgRatePromCollector{collector: promCollector, promMetricsToValue: promMetricsToValue}
	avgRage, err := newAvgRate(a.count, samplingWindow, period, unit)
	if err != nil {
		return nil, fmt.Errorf("new avg rate counter: %w", err)
	}
	a.avgRate = avgRage

	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	runJanitor(avgRage, samplingWindow)
	runtime.SetFinalizer(a, stopJanitor)

	return a, nil
}
