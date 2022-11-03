package dmetrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvgRateCounter(t *testing.T) {
	interval := 100 * time.Millisecond
	period := 3 * time.Second

	counter := NewAtomicCounter()
	r, err := NewAvgRateCounter(counter, interval, period, "blocks")
	require.NoError(t, err)

	counter.Add(1)
	time.Sleep(300 * time.Millisecond)
	counter.Add(1)
	time.Sleep(200 * time.Millisecond)
	counter.Add(1)
	time.Sleep(500 * time.Millisecond)
	counter.Add(1)
	assert.InDelta(t, 0.3, r.RateFloat64(), 0.1)
	assert.Equal(t, "0.3 blocks/100ms (3 total)", r.String())
}
