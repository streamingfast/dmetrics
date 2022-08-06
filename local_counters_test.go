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
			assert.Equal(t, tt.want, NewPerSecondLocalRateCounter(tt.args.unit))
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
			NewPerSecondLocalRateCounter("items"),
			rateGenerator(1*time.Second, 10),
			"10 items/s (10 total)",
		},

		{
			"average standard rate counter",
			NewAvgLocalRateCounter(1*time.Second, "items"),
			avgRateGenerator(125, 150, 175),
			"~150 items/s (450 total)",
		},

		{
			"average standard unit is actually a time unit counter",
			NewAvgLocalRateCounter(1*time.Second, "ms/block"),
			avgRateGenerator(125, 150, 175),
			"~150ms/block (over 1s)",
		},

		{
			"average resets after first round",
			NewAvgLocalRateCounter(1*time.Second, "ms/block"),
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
