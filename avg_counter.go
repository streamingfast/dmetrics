package dmetrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type AvgCounter struct {
	counter        *ratecounter.AvgRateCounter
	samplingWindow time.Duration
	eventType      string
	total          uint64
}

// NewAvgCounter allows you to get the average of an event over the period of time.
// For example, if you want to know the average cache hits in a given time
//
// # Over 1 second, you will increment the average by the number of cache hits
//
// ```
// counter := NewAvgCounter(1*time.Second, "cache hits")
// counter.IncBy(283)
// counter.IncBy(23)
// counter.IncBy(192)
// counter.IncBy(392)
//
// counter.String() == avg 222.5 cache hits (over 1s) [12321 total]
// ```
func NewAvgCounter(samplingWindow time.Duration, eventType string) *AvgCounter {
	return &AvgCounter{
		counter:        ratecounter.NewAvgRateCounter(samplingWindow),
		samplingWindow: samplingWindow,
		eventType:      eventType,
		total:          0,
	}
}

// IncBy adds multiple events (useful for debouncing event counts)
func (c *AvgCounter) IncBy(value int64) {
	if value <= 0 {
		return
	}

	c.counter.Incr(value)
	c.total += uint64(value)
}

func (c *AvgCounter) Average() float64 {
	return c.counter.Rate()
}

func (c *AvgCounter) Total() uint64 {
	return c.total
}

func (c *AvgCounter) AverageString() string {
	return strconv.FormatFloat(c.Average(), 'f', -1, 64)
}

func (c *AvgCounter) String() string {
	return fmt.Sprintf("avg %s %s (in the last %s) [total %d]", c.AverageString(), c.eventType, samplingWindowToString(c.samplingWindow), c.Total())
}

func samplingWindowToString(sampling time.Duration) string {
	switch sampling {
	case 1 * time.Second:
		return "1s"
	case 1 * time.Minute:
		return "1min"
	case 1 * time.Millisecond:
		return "1ms"
	default:
		return sampling.String()
	}
}
