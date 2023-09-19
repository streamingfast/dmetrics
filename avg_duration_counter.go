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
// As an example, if you want to know the average block process time.
//
// Example: if over 1 second you process 3 blocks where the processing
// time respectively takes 2s, 5s, 300ms. The average will yield a result of 2.43s per block.
//
// ```
// counter := NewAvgDurationCounter(30*time.Second, time.Second, "per block")
// counter.AddDuration(2 * time.Second)
// counter.AddDuration(5 * time.Second)
// counter.AddDuration(300 * time.Millisecond)
//
// counter.String() == 2.43s per block (avg over 30s)
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
	return strconv.FormatFloat(c.Average(), 'f', 2, 64)
}

func (c *AvgDurationCounter) String() string {
	return fmt.Sprintf("%s%s %s (avg over %s)", c.AverageString(), timeUnitToString(c.unit), c.description, samplingWindowToString(c.samplingWindow))
}
