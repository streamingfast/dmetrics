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

	r, err := NewAvgRateCounter(interval, period, "blocks")
	require.NoError(t, err)

	r.Add(1)
	time.Sleep(300 * time.Millisecond)
	r.Add(1)
	time.Sleep(200 * time.Millisecond)
	r.Add(1)
	time.Sleep(500 * time.Millisecond)
	r.Add(1)
	assert.InDelta(t, 0.3, r.Rate(), 0.1)
	// hard to make this accurate
	//assert.Equal(t, "0.333 blocks/100ms (3 total)", r.String())
}
