package dmetrics

import (
	"fmt"
	"regexp"
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
	total    uint64

	isAverage bool
}

func NewPerSecondLocalRateCounter(unit string) *LocalCounter {
	return NewLocalRateCounter(1*time.Second, unit)
}

func NewPerMinuteLocalRateCounter(unit string) *LocalCounter {
	return NewLocalRateCounter(1*time.Minute, unit)
}

// NewLocalRateCounter creates a counter on which it's easy to how many time an event happen over a fixed
// period of time.
//
// For example, if over 1 second you process 20 blocks, then querying the counter within this 1s interval
// will yield a result of 20 blocks/s. The rate change as the time moves.
//
// ```
// counter := NewLocalRateCounter(1*time.Second, "s", "blocks")
// counter.IncByElapsed(since1)
// counter.IncByElapsed(since2)
// counter.IncByElapsed(since3)
//
// counter.String() == ~150ms/block (over 1s)
// ```
func NewLocalRateCounter(interval time.Duration, unit string) *LocalCounter {
	return &LocalCounter{(*wrappedCounter)(ratecounter.NewRateCounter(interval)), interval, unit, 0, false}
}

func NewAvgPerSecondLocalRateCounter(unit string) *LocalCounter {
	return NewAvgLocalRateCounter(1*time.Second, unit)
}

func NewAvgPerMinuteLocalRateCounter(unit string) *LocalCounter {
	return NewAvgLocalRateCounter(1*time.Minute, unit)
}

// NewAvgLocalRateCounter creates a counter on which it's easy to get the average of something
// over the period of time. For example, if you want to know the average time a repeated task
// took over the period.
//
// Over 1 second, you process 20 blocks. Each of this block had a different decoding time, by incrementing
// the counter by time elapsed for decoding, you get after the period an average time of decoding for N
// elements.
//
// ```
// counter := NewAvgLocalRateCounter(1*time.Second, "block", "ms")
// counter.IncByElapsed(since1)
// counter.IncByElapsed(since2)
// counter.IncByElapsed(since3)
//
// counter.String() == ~150ms/block (over 1s)
// ```
func NewAvgLocalRateCounter(interval time.Duration, unit string) *LocalCounter {
	return &LocalCounter{ratecounter.NewAvgRateCounter(interval), interval, unit, 0, true}
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

var elapsedPerElementUnitRegex = regexp.MustCompile("^(h|min|s|ms)/.+$")

func (c *LocalCounter) String() string {
	// We perform special handling of composed elemnt with time elapsed per unit like
	// `150ms/block`.
	isElapsedPerElementUnit := elapsedPerElementUnitRegex.MatchString(c.unit)

	if c.isAverage {
		if isElapsedPerElementUnit {
			return fmt.Sprintf("%s%s (over %s)", c.RateString(), c.unit, c.intervalString())
		}

		return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, c.timeUnit(), c.total)
	}

	return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, c.timeUnit(), c.total)
}

func (c *LocalCounter) timeUnit() string {
	switch c.interval {
	case 1 * time.Second:
		return "s"
	case 1 * time.Minute:
		return "min"
	case 1 * time.Millisecond:
		return "ms"
	default:
		return c.interval.String()
	}
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
