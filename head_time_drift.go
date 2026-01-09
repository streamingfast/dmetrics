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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

var headTimeDriftGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "head_block_time_drift",
	Help: "Number of seconds away from real-time",
}, []string{"app"})

var headBlockNumber = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "head_block_number",
}, []string{"app"})

var headBlockRelativeTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "head_block_relative_time",
	Help:    "Number of seconds between real-time and block time, for each block",
	Buckets: []float64{0.1, 0.3, 0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 5, 10},
}, []string{"app"})

type HeadTimeDrift struct {
	headBlockTimeCh chan time.Time
	service         string
	started         *atomic.Bool
}

func (s *Set) NewHeadTimeDrift(service string) *HeadTimeDrift {
	headBlockTimeCh := make(chan time.Time)

	h := &HeadTimeDrift{
		headBlockTimeCh: headBlockTimeCh,
		service:         service,
		started:         atomic.NewBool(false),
	}

	return h
}

func (h *HeadTimeDrift) SetBlockTime(blockTime time.Time) {
	if !h.started.Load() {
		go func() {
			headBlockTime := time.Time{}
			for {
				select {
				case blockTime := <-h.headBlockTimeCh:
					headBlockTime = blockTime
				case <-time.After(500 * time.Millisecond):
				}
				headTimeDriftGauge.WithLabelValues(h.service).Set(time.Since(headBlockTime).Seconds())
			}
		}()
		h.started.Store(true)
	}
	h.headBlockTimeCh <- blockTime
}

// Collect implements prometheus.Collector.
func (h *HeadTimeDrift) Collect(ch chan<- prometheus.Metric) {
	headTimeDriftGauge.Collect(ch)
}

// Describe implements prometheus.Collector.
func (h *HeadTimeDrift) Describe(ch chan<- *prometheus.Desc) {
	headTimeDriftGauge.Describe(ch)
}

func (s *Set) NewHeadBlockNumber(service string) *HeadBlockNum {
	return &HeadBlockNum{
		service: service,
	}
}

var _ prometheus.Collector = (*HeadBlockNum)(nil)

type HeadBlockNum struct {
	service string
}

func (h *HeadBlockNum) SetUint64(blockNum uint64) {
	headBlockNumber.WithLabelValues(h.service).Set(float64(blockNum))
}

// Collect implements prometheus.Collector.
func (h *HeadBlockNum) Collect(ch chan<- prometheus.Metric) {
	headBlockNumber.Collect(ch)
}

// Describe implements prometheus.Collector.
func (h *HeadBlockNum) Describe(ch chan<- *prometheus.Desc) {
	headBlockNumber.Describe(ch)
}

func (s *Set) NewHeadBlockRelativeTime(service string) *HeadBlockRelativeTime {
	return &HeadBlockRelativeTime{
		service: service,
	}
}

var _ prometheus.Collector = (*HeadBlockRelativeTime)(nil)

type HeadBlockRelativeTime struct {
	service string
}

func (h *HeadBlockRelativeTime) SetLastBlock(blockTime time.Time) {
	headBlockNumber.WithLabelValues(h.service).Set(time.Since(blockTime).Seconds())
}

// Collect implements prometheus.Collector.
func (h *HeadBlockRelativeTime) Collect(ch chan<- prometheus.Metric) {
	headBlockNumber.Collect(ch)
}

// Describe implements prometheus.Collector.
func (h *HeadBlockRelativeTime) Describe(ch chan<- *prometheus.Desc) {
	headBlockNumber.Describe(ch)
}

func init() {
	PrometheusRegister(headTimeDriftGauge)
	PrometheusRegister(headBlockNumber)
	PrometheusRegister(headBlockRelativeTime)
}
