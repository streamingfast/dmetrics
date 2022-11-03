package dmetrics

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPerSecondLocalCounter(t *testing.T) {
	type args struct {
		unit string
	}
	tests := []struct {
		name string
		args args
		want *RateCounter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewPerSecondLocalRateCounter(tt.args.unit))
		})
	}
}

func TestRateCounter(t *testing.T) {
	t.Skip("this is more of an example, but usefull to understand")
	interval := 1 * time.Second
	r := NewRateCounter(interval, "event")

	// this will increment the rate counter and wait 200 ms.
	// up until we cross the 1-second mark, the rate increase linearly
	// after the 1-second mark  rate will stabilize, since it will increment
	// events at a constant rate
	for i := 0; i < 10; i++ {
		r.Inc()
		time.Sleep(200 * time.Millisecond)
		fmt.Println(r.String())
	}
}

func TestRateCounter_GoldenPath(t *testing.T) {
	interval := 1 * time.Second
	r := NewRateCounter(interval, "event")

	r.Inc()
	r.Inc()
	r.Inc()
	r.Inc()
	r.Inc()
	assert.EqualValues(t, 5, r.Rate())
	assert.Equal(t, "5 event/s (5 total)", r.String())
}

func TestRateCounter_ExceedsIntervalWindow(t *testing.T) {
	interval := 1 * time.Millisecond
	r := NewRateCounter(interval, "event")

	r.Inc()
	r.Inc()
	r.Inc()
	r.Inc()
	r.Inc()
	time.Sleep(10 * interval)
	assert.EqualValues(t, 0, r.Rate())
	assert.Equal(t, "0 event/ms (5 total)", r.String())
}
