package dmetrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type AvgDurationCounter struct {
	counter        *ratecounter.AvgRateCounter
	samplingWindow time.Duration
	unit           time.Duration
	total          int64
	description    string
}

// NewAvgDurationCounter allows you to get teh average elapsed time of a given process
//. For example, if you want to know the average block process time
//
// For example, if over 1 second you process 3 blocks where the processing
// time respectively takes 2s, 5s, 300ms. Teh average will yield a result of 2.433333333333333s per block.
//
// ```
// counter := NewAvgDurationCounter(1*time.Second, "cache hits", "block")
// counter.AddDuration(2 * time.Second)
// counter.AddDuration(5 * time.Second)
// counter.AddDuration(300 * time.Millisecond)
//
// counter.String() == avg 2.4333333333333336s block (in the last 1s) [total 7.3s]
// ```

func NewAvgDurationCounter(samplingWindow time.Duration, unit time.Duration, description string) *AvgDurationCounter {
	return &AvgDurationCounter{
		counter:        ratecounter.NewAvgRateCounter(samplingWindow),
		samplingWindow: samplingWindow,
		unit:           unit,
		total:          0,
		description:    description,
	}
}

func (c *AvgDurationCounter) AddElapsedTime(start time.Time) {
	elapsed := time.Since(start)
	if elapsed <= 0 {
		return
	}
	c.AddDuration(elapsed)
}

func (c *AvgDurationCounter) AddDuration(dur time.Duration) {
	c.counter.Incr(int64(dur))
	c.total += int64(dur)
}

func (c *AvgDurationCounter) Average() float64 {
	return c.counter.Rate() / float64(c.unit)
}

func (c *AvgDurationCounter) Total() float64 {
	return float64(c.total) / float64(c.unit)
}

func (c *AvgDurationCounter) AverageString() string {
	return strconv.FormatFloat(c.Average(), 'f', -1, 64)
}

func (c *AvgDurationCounter) String() string {
	total := strconv.FormatFloat(c.Total(), 'f', -1, 64)
	unitStr := timeUnitToString(c.unit)
	return fmt.Sprintf("avg %s%s %s (in the last %s) [total %s%s]", c.AverageString(), unitStr, c.description, samplingWindowToString(c.samplingWindow), total, unitStr)
}

func (c *AvgDurationCounter) intervalString() string {
	switch c.samplingWindow {
	case 1 * time.Second:
		return "1s"
	case 1 * time.Minute:
		return "1min"
	case 1 * time.Millisecond:
		return "1ms"
	default:
		return c.samplingWindow.String()
	}
}
