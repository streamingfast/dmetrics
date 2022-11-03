package dmetrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalAvgCounter_SameValue(t *testing.T) {
	interval := 500 * time.Millisecond
	r := NewAvgCounter(interval, "cache hits")

	r.IncBy(1)
	r.IncBy(1)
	r.IncBy(1)
	assert.EqualValues(t, 1, r.Average())
	assert.Equal(t, "avg 1 cache hits (in the last 500ms) [total 3]", r.String())

	time.Sleep(2 * interval)

	r.IncBy(283)
	r.IncBy(23)
	r.IncBy(192)
	r.IncBy(392)
	expectedAvg := float64(283+23+192+392) / float64(4)
	assert.InDelta(t, expectedAvg, r.Average(), 0)
	assert.Equal(t, "avg 222.5 cache hits (in the last 500ms) [total 893]", r.String())
}
