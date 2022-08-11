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
	"github.com/prometheus/client_golang/prometheus"
)

var appReady = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "ready",
	Help: "readiness of an app. 1 if ready, 0 otherwise",
},
	[]string{"app"})

type AppReadiness struct {
	service string
}

func (s *Set) NewAppReadiness(service string) *AppReadiness {
	a := &AppReadiness{
		service: service,
	}
	a.SetNotReady()
	return a
}

func (a *AppReadiness) SetReady() {
	appReady.WithLabelValues(a.service).Set(1)
}

func (a *AppReadiness) SetNotReady() {
	appReady.WithLabelValues(a.service).Set(0)
}

func init() {
	PrometheusRegister(appReady)
}
