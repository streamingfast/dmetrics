package dmetrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type wrappedCounter ratecounter.RateCounter

func (c *wrappedCounter) Incr(val int64) {
	(*ratecounter.RateCounter)(c).Incr(val)
}

func (c *wrappedCounter) Rate() float64 {
	return float64((*ratecounter.RateCounter)(c).Rate())
}

type counter interface {
	Incr(val int64)
	Rate() float64
}

type LocalCounter struct {
	counter  counter
	interval time.Duration
	unit     string
	timeUnit string
	total    uint64

	isAverage bool
}

func NewPerSecondLocalCounter(unit string) *LocalCounter {
	return NewLocalCounter(1*time.Second, "s", unit)
}

func NewPerMinuteLocalCounter(unit string) *LocalCounter {
	return NewLocalCounter(1*time.Minute, "min", unit)
}

func NewLocalCounter(interval time.Duration, timeUnit string, unit string) *LocalCounter {
	return &LocalCounter{(*wrappedCounter)(ratecounter.NewRateCounter(interval)), interval, unit, timeUnit, 0, false}
}

func NewAvgPerSecondLocalCounter(unit string) *LocalCounter {
	return NewAvgLocalCounter(1*time.Second, "s", unit)
}

func NewAvgPerMinuteLocalCounter(unit string) *LocalCounter {
	return NewAvgLocalCounter(1*time.Minute, "min", unit)
}

// NewAvgLocalCounter creates a counter on which it's easy to get the average of something
// over the period of time. For example, if you want to know the average time a repeated task
// took over the period.
//
// Over 1 second, you process 20 blocks. Each of this block had a different decoding time, by incrementing
// the counter by time elapsed for decoding, you get after the period an average time of decoding for N
// elements.
//
// ```
// counter := NewAvgLocalCounter(1*time.Second, "block", "ms")
// counter.IncByElapsed(since1)
// counter.IncByElapsed(since2)
// counter.IncByElapsed(since3)
//
// counter.String() == ~150ms/block (over 1s)
// ```
func NewAvgLocalCounter(interval time.Duration, timeUnit string, unit string) *LocalCounter {
	return &LocalCounter{ratecounter.NewAvgRateCounter(interval), interval, unit, timeUnit, 0, true}
}

// Incr add 1 event into the RateCounter
func (c *LocalCounter) Inc() {
	c.counter.Incr(1)
	c.total++
}

func (c *LocalCounter) IncBy(value int64) {
	if value <= 0 {
		return
	}

	c.counter.Incr(value)
	c.total += uint64(value)
}

func (c *LocalCounter) IncByElapsedTime(start time.Time) {
	elapsed := time.Since(start)
	if elapsed <= 0 {
		return
	}

	c.counter.Incr(int64(elapsed))
	c.total += uint64(elapsed)
}

func (c *LocalCounter) Total() uint64 {
	return c.total
}

func (c *LocalCounter) RateInt64() int64 {
	return int64(c.counter.Rate())
}

func (c *LocalCounter) RateFloat64() float64 {
	return c.counter.Rate()
}

func (c *LocalCounter) RateString() string {
	if !c.isAverage {
		return strconv.FormatInt(c.RateInt64(), 10)
	}

	return strconv.FormatFloat(c.RateFloat64(), 'f', -1, 64)
}

func (c *LocalCounter) String() string {
	// For what looks like time unit, we put the unit directly after the rate because
	// `150ms/block` than `150 ms/block`.
	isUnitAsTimeUnit := c.unit == "h" || c.unit == "min" || c.unit == "s" || c.unit == "ms"

	spaceAfterRate := " "
	if isUnitAsTimeUnit {
		spaceAfterRate = ""
	}

	if c.isAverage {
		rate := c.RateString()
		if rate != "0" {
			rate = "~" + rate
		}

		return fmt.Sprintf("%s%s%s/%s (over %s)", rate, spaceAfterRate, c.unit, c.timeUnit, c.intervalString())
	}

	total := fmt.Sprintf("%d total", c.total)
	if isUnitAsTimeUnit {
		total = fmt.Sprintf("%d%s total", c.total, c.unit)
	}

	return fmt.Sprintf("%s%s%s/%s (%s)", c.RateString(), spaceAfterRate, c.unit, c.timeUnit, total)
}

func (c *LocalCounter) intervalString() string {
	switch c.interval {
	case 1 * time.Second:
		return "1s"
	case 1 * time.Minute:
		return "1min"
	case 1 * time.Millisecond:
		return "1ms"
	default:
		return c.interval.String()
	}
}
