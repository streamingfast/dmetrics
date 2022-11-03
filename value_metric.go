package dmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// ValueFromMetric can be used to extract the value of a Prometheus Metric object.
//
// *Important* This for now does not correctly handles `Vec` like metrics since the
// actual logic is to return the first ever value encountered while in a `Vec` metric,
// there is usually N values, one per label.
type ValueFromMetric struct {
	metric prometheus.Collector
	unit   string
}

func NewValueFromMetric(metric prometheus.Collector, unit string) *ValueFromMetric {
	return &ValueFromMetric{metric, unit}
}

func (c *ValueFromMetric) ValueUint() uint64 {
	return uint64(c.ValueFloat())
}

func (c *ValueFromMetric) ValueFloat() float64 {
	metricChan := make(chan prometheus.Metric, 16)
	go func() {
		c.metric.Collect(metricChan)
		close(metricChan)
	}()

	var firstValue *float64
	for value := range metricChan {
		if firstValue != nil {
			// We must fully consume the metric chan, so if the first value has been found, discard any more reading
			continue
		}

		model := new(dto.Metric)
		err := value.Write(model)
		if err != nil {
			panic(err)
		}

		if model.Gauge != nil && model.Gauge.Value != nil {
			firstValue = model.Gauge.Value
		}
	}

	if firstValue != nil {
		return *firstValue
	}

	return 0.0
}
