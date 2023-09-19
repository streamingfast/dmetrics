package dmetrics

import (
	"fmt"
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
//
// The `unit` parameter can be 0, in which case the unit will be inferred based on the
// actual duration, e.g. if the average is 1.5s, the unit will be 1s while if the average is
// 10us, the unit will be 10us.
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

func (c *AvgDurationCounter) Average() time.Duration {
	return time.Duration(c.counter.Rate())
}

func (c *AvgDurationCounter) Total() time.Duration {
	return time.Duration(c.total)
}

func (c *AvgDurationCounter) AverageString() string {
	return durationToString(time.Duration(c.Average()), c.unit)
}

func (c *AvgDurationCounter) String() string {
	return fmt.Sprintf("%s %s (avg over %s)", c.AverageString(), c.description, samplingWindowToString(c.samplingWindow))
}
