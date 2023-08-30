package dmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// ValuesFromMetric can be used to extract the values of a Prometheus Metric 'Vec' object, i.e. with a 'label' dimension.
type ValuesFromMetric struct {
	metric prometheus.Collector
}

func NewValuesFromMetric(metric prometheus.Collector) *ValuesFromMetric {
	return &ValuesFromMetric{metric}
}

// Uints(label string) converts values to uint64 of the Floats(label string) function
func (c *ValuesFromMetric) Uints(label string) map[string]uint64 {
	out := make(map[string]uint64)
	for k, v := range c.Floats(label) {
		out[k] = uint64(v)
	}

	return out
}

// Floats(label string) gets you the float64 values for each value of the given label.
// Values without that label or with a nil value for that label are discarded
func (c *ValuesFromMetric) Floats(label string) map[string]float64 {
	metricChan := make(chan prometheus.Metric, 16)
	go func() {
		c.metric.Collect(metricChan)
		close(metricChan)
	}()

	out := make(map[string]float64)
	for value := range metricChan {

		model := new(dto.Metric)
		err := value.Write(model)
		if err != nil {
			panic(err)
		}

		var labelValue *string
		for _, pair := range model.Label {
			if pair.Name != nil && *pair.Name == label && pair.Value != nil {
				labelValue = pair.Value
				break
			}
		}
		if labelValue == nil {
			continue
		}
		if _, ok := out[*labelValue]; ok {
			// We must fully consume the metric chan, so if the first value has been found, discard any more reading
			continue
		}

		if model.Gauge != nil && model.Gauge.Value != nil {
			out[*labelValue] = *model.Gauge.Value
		}
	}

	return out
}
