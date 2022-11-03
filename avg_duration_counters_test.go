package dmetrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAvgDurationCounter(t *testing.T) {
	samplingWindow := 1 * time.Second
	r := NewAvgDurationCounter(samplingWindow, time.Second, "block")

	r.AddDuration(2 * time.Second)
	r.AddDuration(5 * time.Second)
	r.AddDuration(300 * time.Millisecond)
	// (2 + 5 + 0.5) / 3 == 2.4333
	assert.InDelta(t, 2.433333333333333, r.Average(), 0.1)
	assert.Equal(t, "avg 2.4333333333333336s block (in the last 1s) [total 7.3s]", r.String())
}
