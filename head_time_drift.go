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

func (s *Set) NewHeadBlockNumber(service string) *HeadBlockNum {
	return &HeadBlockNum{
		service: service,
	}
}

type HeadBlockNum struct {
	service string
}

func (h *HeadBlockNum) SetUint64(blockNum uint64) {
	headBlockNumber.WithLabelValues(h.service).Set(float64(blockNum))
}

func init() {
	PrometheusRegister(headTimeDriftGauge)
	PrometheusRegister(headBlockNumber)
}
