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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestSet_Plain(t *testing.T) {
	collector := hookTestRegister()

	set := NewSet()
	set.NewGauge("test1")
	set.NewGaugeVec("test2", []string{})
	set.NewCounter("test3")
	set.NewCounterVec("test4", []string{})
	set.NewHistogram("test5")
	set.NewHistogramVec("test6", []string{})
	set.NewHeadTimeDrift("service7")

	assert.Equal(t, 0, collector.count())
	assert.Equal(t, 6, len(set.metrics)) // excludes `HeadTimeDrift`

	set.Register()
	assert.Equal(t, 6, collector.count())
	assert.Equal(t, 6, len(set.metrics))
}

func TestSet_Register_Idempotent(t *testing.T) {
	collector := hookTestRegister()

	set := NewSet()
	set.NewGauge("test1")

	set.Register()
	set.Register()
	assert.Equal(t, 1, collector.count())
	assert.Equal(t, 1, len(set.metrics))
}

func TestMetric_SanitizeName(t *testing.T) {
	hookTestRegister()

	tests := []struct {
		in       string
		expected string
	}{
		{"test1 space", `Desc{fqName: "test1_space", help: "h", constLabels: {}, variableLabels: []}`},
		{"test1-space", `Desc{fqName: "test1_space", help: "h", constLabels: {}, variableLabels: []}`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			set := NewSet()
			gauge := set.NewGauge(test.in, "h")

			assert.Equal(t, test.expected, gauge.Native().Desc().String())
		})
	}
}

func TestMetric_Help(t *testing.T) {
	hookTestRegister()

	tests := []struct {
		name     string
		help     []string
		expected string
	}{
		{"a-b", []string{}, `Desc{fqName: "a_b", help: "A B", constLabels: {}, variableLabels: []}`},
		{"a-b", []string{""}, `Desc{fqName: "a_b", help: "A B", constLabels: {}, variableLabels: []}`},
		{"a-b", []string{"test"}, `Desc{fqName: "a_b", help: "test", constLabels: {}, variableLabels: []}`},
		{"a-b", []string{"%s test"}, `Desc{fqName: "a_b", help: "A B test", constLabels: {}, variableLabels: []}`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			set := NewSet()
			gauge := set.NewGauge(test.name, test.help...)

			assert.Equal(t, test.expected, gauge.Native().Desc().String())
		})
	}
}

func TestSet_WithPrefix(t *testing.T) {
	set := NewSet(PrefixNameWith("prefix"))
	gauge := set.NewGauge("test space", "%s are", "multiple")

	expectedDesc := `Desc{fqName: "prefix_test_space", help: "Prefix Test Space are multiple", constLabels: {}, variableLabels: []}`

	assert.Equal(t, expectedDesc, gauge.Native().Desc().String())
}

type registerController struct {
	collectedMetrics []prometheus.Collector
}

func (c *registerController) register(collectors ...prometheus.Collector) {
	c.collectedMetrics = append(c.collectedMetrics, collectors...)
}

func (c *registerController) count() int {
	return len(c.collectedMetrics)
}

func hookTestRegister() *registerController {
	collector := &registerController{}
	PrometheusRegister = collector.register

	return collector
}
