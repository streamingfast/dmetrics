package dmetrics

import (
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
		want *LocalCounter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewPerSecondLocalCounter(tt.args.unit))
		})
	}
}

func TestLocalCounter_String(t *testing.T) {
	tests := []struct {
		name      string
		counter   *LocalCounter
		generator func(*LocalCounter)
		want      string
	}{
		{
			"standard rate counter",
			NewPerSecondLocalCounter("items"),
			rateGenerator(1*time.Second, 10),
			"10 items/s (10 total)",
		},

		{
			"unit is actually a time unit rate counter",
			NewPerSecondLocalCounter("ms"),
			rateGenerator(1*time.Second, 10),
			"10ms/s (10ms total)",
		},

		{
			"average standard rate counter",
			NewAvgLocalCounter(1*time.Second, "s", "items"),
			avgRateGenerator(125, 150, 175),
			"~150 items/s (over 1s)",
		},

		{
			"average standard unit is actually a time unit counter",
			NewAvgLocalCounter(1*time.Second, "block", "ms"),
			avgRateGenerator(125, 150, 175),
			"~150ms/block (over 1s)",
		},

		{
			"average resets after first round",
			NewAvgLocalCounter(1*time.Second, "block", "ms"),
			avgRateGenerator(125, 150, 175, sleep(1*time.Second)),
			"0ms/block (over 1s)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.generator(tt.counter)

			assert.Equal(t, tt.want, tt.counter.String())
		})
	}
}

func rateGenerator(period time.Duration, elements ...interface{}) func(counter *LocalCounter) {
	return func(counter *LocalCounter) {
		for _, element := range elements {
			switch v := element.(type) {
			case uint:
				counter.IncBy(int64(v))
			case int:
				counter.IncBy(int64(v))
			case uint64:
				counter.IncBy(int64(v))
			case int64:
				counter.IncBy(v)
			case sleep:
				time.Sleep(time.Duration(v))
			}
		}
	}
}

type sleep int64

func avgRateGenerator(elements ...interface{}) func(counter *LocalCounter) {
	return func(counter *LocalCounter) {
		for _, element := range elements {
			switch v := element.(type) {
			case uint:
				counter.IncBy(int64(v))
			case int:
				counter.IncBy(int64(v))
			case uint64:
				counter.IncBy(int64(v))
			case int64:
				counter.IncBy(v)
			case sleep:
				time.Sleep(time.Duration(v))
			}
		}
	}
}
