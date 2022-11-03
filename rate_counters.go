package dmetrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type RateCounter struct {
	counter  *ratecounter.RateCounter
	interval time.Duration
	unit     string
	total    uint64
}

// NewRateCounter allows you to know  how many times an event happen over a fixed period of time
//
// For example, if over 1 second you process 20 blocks, then querying the counter within this 1s samplingWindow
// will yield a result of 20 blocks/s. The rate change as the time moves.
//
// ```
// counter := NewRateCounter(1*time.Second, "s", "blocks")
// counter.Incr()
// counter.Incr()
// counter.Incr()
//
// counter.String() == 3 blocks/s (over 1s)
//
// IMPORTANT: The rate is calculated by the number of events that occurred in the last sampling window
// thus if from time 0s to 1s you have 3 events the rate is 3 events/sec; subsequently if you wait another 2
// seconds without any event occurring, then query the rate you will get 0 event/sec. If you want an Avg rate over
// multiple sampling window Use AvgRateCounter
// ```
func NewRateCounter(interval time.Duration, unit string) *RateCounter {
	return &RateCounter{
		counter:  ratecounter.NewRateCounter(interval),
		interval: interval,
		unit:     unit,
		total:    0,
	}
}

func NewPerSecondLocalRateCounter(unit string) *RateCounter {
	return NewRateCounter(1*time.Second, unit)
}

func NewPerMinuteLocalRateCounter(unit string) *RateCounter {
	return NewRateCounter(1*time.Minute, unit)
}

// Incr add 1 event into the RateCounter
func (c *RateCounter) Inc() {
	c.counter.Incr(1)
	c.total++
}

// IncrBy adds multiple events inot the RateCounter
func (c *RateCounter) IncBy(value int64) {
	if value <= 0 {
		return
	}

	c.counter.Incr(value)
	c.total += uint64(value)
}

func (c *RateCounter) Total() uint64 {
	return c.total
}

func (c *RateCounter) Rate() int64 {
	return c.counter.Rate()
}

func (c *RateCounter) RateString() string {
	return strconv.FormatInt(c.counter.Rate(), 10)
}

//var ratioUnitRegex = regexp.MustCompile("^[^/]+/.+$")
//var elapsedPerElementUnitPrefixRegex = regexp.MustCompile("^(h|min|s|ms)/")

func (c *RateCounter) String() string {
	return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, timeUnitToString(c.interval), c.total)
}
