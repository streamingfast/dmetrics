package dmetrics

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/streamingfast/dmetrics/ring"
	"go.uber.org/atomic"
)

// avgRateCounter exists because we use a janitor that is invoked when the object embedding `avgRate
// is actually gargabed collected. This object is of different type and we cannot know of which this is
// going to be.
//
// So we have defined this interface, that `avgRate` implements itself so everyone that embeds `avgRate`
// struct is automatically a `avgRateCounter` interface. In the `janitor` finalizer, we then use this
// interface to terminate the work.
type avgRateCounter interface {
	getAvgRate() *avgRate
}

var _ avgRateCounter = (*AvgRateCounter)(nil)

type AvgRateCounter struct {
	*avgRate
	c *atomic.Uint64
}

func MustNewAvgRateCounter(samplingWindow time.Duration, period time.Duration, unit string) *AvgRateCounter {
	a, err := NewAvgRateCounter(samplingWindow, period, unit)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAvgRateCounter Tracks the average rate over a period of time. The rate is
// computed by accumulating the total count at the <samplingWindow> interval and
// averaging them our ove the number of <period> defined.
// Suppose there is a block count that increments as follows
//
//	0s to 1s -> 10 blocks
//	1s to 2s -> 3 blocks
//	2s to 3s -> 0 blocks
//	3s to 4s -> 7 blocks
//
// If your  samplingWindow = 1s and your period = 4s, the rate will be computed as
//
//	(10 + 3 + 0 + 7)/4 = 5 blocks/sec
//
// If your  samplingWindow = 1s and your period = 3s, the rate will be computed as
//
//	(10 + 3 + 0)/4 = 4.33 blocks/sec
//
// then when the "window moves" you would get
//
//	(3 + 0 + 7)/4 = 3.333 blocks/sec
func NewAvgRateCounter(samplingWindow time.Duration, period time.Duration, unit string) (*AvgRateCounter, error) {
	a := &AvgRateCounter{c: atomic.NewUint64(0)}
	avgRage, err := newAvgRate(a.count, samplingWindow, period, unit)
	if err != nil {
		return nil, fmt.Errorf("new avg rate counter: %w", err)
	}
	a.avgRate = avgRage

	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	runJanitor(avgRage, samplingWindow)
	runtime.SetFinalizer(a, stopJanitor)

	return a, nil
}

// Add tracks a number of events, to be used to compute the rage
func (a *AvgRateCounter) Add(v uint64) {
	a.c.Add(v)
}

func (a *AvgRateCounter) count() uint64 {
	return a.c.Load()
}

func (c *AvgRateCounter) Stop() {
	stopJanitor(c)
}

type CountableFunc = func() uint64

// NewAvgRateCounter AvgRateCounter can be used to extract the average rate of a Countable object
// over a period of time. The computation is to accumulate the instant metric each <samplingWindow>
// and obtain an average over <period> duration. The <period> must be greater than
// <samplingWindow> and should be a multiple of it (enforced).
//
// If you for example want to log the rate of something each 30s and the rate is checked each 1s,
// your <period> should be set to 30s.

type avgRate struct {
	//counter        prometheus.Collector
	counterFunc    CountableFunc
	samplingWindow time.Duration
	unit           string
	bucketCount    uint64
	janitor        *janitor

	totals      *ring.Ring[uint64]
	actualTotal uint64
	actualCount uint64
}

func newAvgRate(counter CountableFunc, samplingWindow time.Duration, period time.Duration, unit string) (*avgRate, error) {
	if samplingWindow == 0 {
		return nil, fmt.Errorf("sampling window must be greater then 0")
	}

	if samplingWindow > period {
		return nil, fmt.Errorf("interval (%s) must be lower than averageTime (%s) but it's not", samplingWindow, period)
	}

	if period%samplingWindow != 0 {
		return nil, fmt.Errorf("averageTime (%s) must be divisible by samplingWindow (%s) without a remainder but it's not", period, samplingWindow)
	}

	bucketCount := (uint64(period / samplingWindow)) + 1

	return &avgRate{
		counterFunc:    counter,
		samplingWindow: samplingWindow,
		unit:           unit,
		bucketCount:    bucketCount,
		totals:         ring.New[uint64](int(bucketCount)),
	}, nil
}

func (c *avgRate) Total() uint64      { return c.actualTotal }
func (c *avgRate) Rate() float64      { return c.rate() }
func (c *avgRate) RateString() string { return strconv.FormatFloat(c.Rate(), 'f', 3, 64) }
func (c *avgRate) String() string {
	return fmt.Sprintf("%s %s/%s (%d total)", c.RateString(), c.unit, timeUnitToString(c.samplingWindow), c.Total())
}
func (c *avgRate) rate() float64 {
	skip := uint64(0)
	if c.actualCount < uint64(c.bucketCount) {
		// We do an extra minus one because we are interested about delta and there is always `c.bucketCount - 1` deltas
		skip = c.bucketCount - c.actualCount - 1
	}

	var sum uint64
	var deltaCount uint64
	var valueCount uint64
	var previousData *uint64

	c.totals.Do(func(total uint64) {
		if valueCount > skip && previousData != nil {
			sum += total - *previousData
			deltaCount++
		}

		previousData = &total
		valueCount++
	})

	return float64(sum) / float64(deltaCount)
}

// SyncNow forces a sync to retrieve the value(s) of the counter(s)
// and update the average rate.
//
// This call is blocking until the sync is performed.
func (c *avgRate) SyncNow() {
	if c.janitor != nil {
		select {
		case c.janitor.wake <- true:
		case <-c.janitor.stop:
			return
		}

		// Block until the sync is performed (or janitor is stopped)
		select {
		case <-c.janitor.syncPerformed:
		case <-c.janitor.stop:
		}
	}
}

func (c *avgRate) syncNow() {
	c.actualCount++
	c.actualTotal = c.counterFunc()

	c.totals.Value = c.actualTotal
	c.totals = c.totals.Next()
}

func (c *avgRate) getAvgRate() *avgRate {
	return c
}

type janitor struct {
	samplingWindow time.Duration
	wake           chan bool
	stop           chan bool
	syncPerformed  chan bool
	once           *sync.Once
}

func (j *janitor) run(r *avgRate) {
	ticker := time.NewTicker(j.samplingWindow)

	for {
		select {
		case <-ticker.C:
			r.syncNow()
		case <-j.wake:
			r.syncNow()

			select {
			case j.syncPerformed <- true:
			default:
				// Channel is full, no one consumes, we don't care
			}
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c avgRateCounter) {
	j := c.getAvgRate().janitor

	if j != nil {
		j.once.Do(func() {
			j.stop <- true
			close(j.stop)
			close(j.wake)
		})
	}
}

func runJanitor(r *avgRate, samplingWindow time.Duration) {
	j := &janitor{
		samplingWindow: samplingWindow,
		stop:           make(chan bool, 1),
		wake:           make(chan bool, 1),
		syncPerformed:  make(chan bool, 1),
		once:           &sync.Once{},
	}
	r.janitor = j

	go j.run(r)
}
