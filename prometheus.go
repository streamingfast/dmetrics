// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dmetrics

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var mutex sync.Mutex

var PrometheusRegister = prometheus.MustRegister

var NoOpPrometheusRegister = func(c ...prometheus.Collector) {}

type Set struct {
	autoRegister  bool
	metricsPrefix string

	metrics      []Metric
	isRegistered bool
	parent       *Set
	children     []*Set
}

type Option func(s *Set)

// Register from `metrics.go`
func Register(sets ...*Set) {
	for _, set := range sets {
		set.Register()
	}
}

// PrefixNameWith will prefix all metric of this given set using the
// following prefix.
func PrefixNameWith(prefix string) Option {
	return func(s *Set) {
		s.metricsPrefix = prefix
	}
}

// NewSet creates a set of metrics that can then be used to create
// a varieties of specific metrics (Gauge, Counter, Histogram).
func NewSet(options ...Option) *Set {
	s := &Set{}
	for _, option := range options {
		option(s)
	}

	return s
}

func (s *Set) add(metric Metric) Metric {
	s.metrics = append(s.metrics, metric)

	return metric
}

func (s *Set) Register() {
	mutex.Lock()
	defer mutex.Unlock()

	if s.isRegistered {
		return
	}

	for _, metric := range s.metrics {
		PrometheusRegister(metric)
	}

	s.isRegistered = true
}

type Metric interface {
	prometheus.Collector
}

// Compile checks to ensure our struct implements the proper metrics
var _ prometheus.Collector = (*Gauge)(nil)
var _ prometheus.Collector = (*GaugeVec)(nil)
var _ prometheus.Collector = (*Counter)(nil)
var _ prometheus.Collector = (*CounterVec)(nil)
var _ prometheus.Collector = (*Histogram)(nil)
var _ prometheus.Collector = (*HistogramVec)(nil)

type Gauge struct {
	p prometheus.Gauge
}

func (g *Gauge) Inc()                                { g.p.Inc() }
func (g *Gauge) Dec()                                { g.p.Dec() }
func (g *Gauge) SetUint64(value uint64)              { g.p.Set(float64(value)) }
func (g *Gauge) SetFloat64(value float64)            { g.p.Set(float64(value)) }
func (g *Gauge) Native() prometheus.Gauge            { return g.p }
func (g *Gauge) Describe(in chan<- *prometheus.Desc) { g.p.Describe(in) }
func (g *Gauge) Collect(in chan<- prometheus.Metric) { g.p.Collect(in) }

func (s *Set) NewGauge(name string, helpChunks ...string) *Gauge {
	name = s.computeMetricName(name)
	g := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)})

	return s.add(&Gauge{
		p: g,
	}).(*Gauge)
}

type Counter struct {
	p prometheus.Counter
}

func (c *Counter) Inc()                                { c.p.Inc() }
func (c *Counter) AddInt(value int)                    { c.p.Add(float64(value)) }
func (c *Counter) AddInt64(value int64)                { c.p.Add(float64(value)) }
func (c *Counter) AddUint64(value uint64)              { c.p.Add(float64(value)) }
func (c *Counter) AddFloat64(value float64)            { c.p.Add(float64(value)) }
func (c *Counter) Native() prometheus.Counter          { return c.p }
func (g *Counter) Describe(in chan<- *prometheus.Desc) { g.p.Describe(in) }
func (g *Counter) Collect(in chan<- prometheus.Metric) { g.p.Collect(in) }

func (s *Set) NewCounter(name string, helpChunks ...string) *Counter {
	name = s.computeMetricName(name)
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)})

	return s.add(&Counter{
		p: c,
	}).(*Counter)
}

type CounterVec struct {
	p *prometheus.CounterVec
}

func (c *CounterVec) Inc(labels ...string) { c.p.WithLabelValues(labels...).Inc() }

func (c *CounterVec) AddInt(value int, labels ...string) {
	c.p.WithLabelValues(labels...).Add(float64(value))
}
func (c *CounterVec) AddInt64(value int64, labels ...string) {
	c.p.WithLabelValues(labels...).Add(float64(value))
}
func (c *CounterVec) AddUint64(value uint64, labels ...string) {
	c.p.WithLabelValues(labels...).Add(float64(value))
}
func (c *CounterVec) AddFloat64(value float64, labels ...string) {
	c.p.WithLabelValues(labels...).Add(float64(value))
}
func (c *CounterVec) DeleteLabelValues(labels ...string) {
	c.p.DeleteLabelValues(labels...)
}

func (g *CounterVec) Native() *prometheus.CounterVec      { return g.p }
func (g *CounterVec) Describe(in chan<- *prometheus.Desc) { g.p.Describe(in) }
func (g *CounterVec) Collect(in chan<- prometheus.Metric) { g.p.Collect(in) }

func (s *Set) NewCounterVec(name string, labels []string, helpChunks ...string) *CounterVec {
	name = s.computeMetricName(name)
	c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)}, labels)

	return s.add(&CounterVec{
		p: c,
	}).(*CounterVec)
}

type GaugeVec struct {
	p *prometheus.GaugeVec
}

func (g *GaugeVec) Inc(labels ...string) { g.p.WithLabelValues(labels...).Inc() }

func (g *GaugeVec) Dec(labels ...string) { g.p.WithLabelValues(labels...).Dec() }

func (g *GaugeVec) SetInt(value int, labels ...string) {
	g.p.WithLabelValues(labels...).Set(float64(value))
}

func (g *GaugeVec) SetInt64(value int64, labels ...string) {
	g.p.WithLabelValues(labels...).Set(float64(value))
}

func (g *GaugeVec) SetUint64(value uint64, labels ...string) {
	g.p.WithLabelValues(labels...).Set(float64(value))
}

func (g *GaugeVec) SetFloat64(value float64, labels ...string) {
	g.p.WithLabelValues(labels...).Set(float64(value))
}

func (g *GaugeVec) DeleteLabelValues(labels ...string) {
	g.p.DeleteLabelValues(labels...)
}

func (g *GaugeVec) Native() *prometheus.GaugeVec        { return g.p }
func (g *GaugeVec) Describe(in chan<- *prometheus.Desc) { g.p.Describe(in) }
func (g *GaugeVec) Collect(in chan<- prometheus.Metric) { g.p.Collect(in) }

func (s *Set) NewGaugeVec(name string, labels []string, helpChunks ...string) *GaugeVec {
	name = s.computeMetricName(name)
	g := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)}, labels)

	return s.add(&GaugeVec{
		p: g,
	}).(*GaugeVec)
}

type Histogram struct {
	p prometheus.Histogram
}

func (s *Set) NewHistogram(name string, helpChunks ...string) *Histogram {
	name = s.computeMetricName(name)
	h := prometheus.NewHistogram(prometheus.HistogramOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)})

	return s.add(&Histogram{
		p: h,
	}).(*Histogram)
}

func (h *Histogram) ObserveDuration(value time.Duration) {
	h.p.Observe(value.Seconds())
}

func (h *Histogram) ObserveSince(value time.Time) {
	h.p.Observe(time.Since(value).Seconds())
}

func (h *Histogram) ObserveInt(value int64) {
	h.p.Observe(float64(value))
}

func (h *Histogram) ObserveInt64(value int64) {
	h.p.Observe(float64(value))
}

func (h *Histogram) ObserveUint64(value int64) {
	h.p.Observe(float64(value))
}

func (h *Histogram) ObserveFloat64(value float64) {
	h.p.Observe(value)
}

func (h *Histogram) Native() prometheus.Histogram        { return h.p }
func (h *Histogram) Describe(in chan<- *prometheus.Desc) { h.p.Describe(in) }
func (h *Histogram) Collect(in chan<- prometheus.Metric) { h.p.Collect(in) }

type HistogramVec struct {
	p *prometheus.HistogramVec
}

func (s *Set) NewHistogramVec(name string, labels []string, helpChunks ...string) *HistogramVec {
	name = s.computeMetricName(name)
	h := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: generateMetricsHelp(name, helpChunks)}, labels)

	return s.add(&HistogramVec{
		p: h,
	}).(*HistogramVec)
}

func (h *HistogramVec) ObserveDuration(value time.Duration, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(value.Seconds())
}

func (h *HistogramVec) ObserveSince(value time.Time, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(time.Since(value).Seconds())
}

func (h *HistogramVec) ObserveInt(value int64, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(float64(value))
}

func (h *HistogramVec) ObserveInt64(value int64, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(float64(value))
}

func (h *HistogramVec) ObserveUint64(value int64, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(float64(value))
}

func (h *HistogramVec) ObserveFloat64(value float64, labels ...string) {
	h.p.WithLabelValues(labels...).Observe(value)
}

func (h *HistogramVec) DeleteLabelValues(labels ...string) {
	h.p.DeleteLabelValues(labels...)
}

func (h *HistogramVec) Native() *prometheus.HistogramVec    { return h.p }
func (h *HistogramVec) Describe(in chan<- *prometheus.Desc) { h.p.Describe(in) }
func (h *HistogramVec) Collect(in chan<- prometheus.Metric) { h.p.Collect(in) }

var nameSanitizerRegex = regexp.MustCompile("[^a-zA-Z0-9_]+")

func (s *Set) computeMetricName(in string) string {
	if s.metricsPrefix != "" {
		return s.metricsPrefix + "_" + sanitizeName(in)
	}

	return sanitizeName(in)
}

func sanitizeName(in string) string {
	return nameSanitizerRegex.ReplaceAllLiteralString(in, "_")
}

func generateMetricsHelp(name string, helpChunks []string) string {
	help := strings.Join(helpChunks, " ")
	if help == "" {
		help = "%s"
	}

	if !strings.Contains(help, "%s") {
		return help
	}

	return fmt.Sprintf(help, strings.Title(strings.ReplaceAll(name, "_", " ")))
}
