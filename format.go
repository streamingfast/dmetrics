package dmetrics

import (
	"fmt"
	"strconv"
	"time"
)

var InferUnit = time.Duration(0)

func timeUnitToString(unit time.Duration) string {
	switch unit {
	case 1 * time.Second:
		return "s"
	case 1 * time.Minute:
		return "min"
	case 1 * time.Millisecond:
		return "ms"
	default:
		return unit.String()
	}
}

func durationToString(d time.Duration, unit time.Duration) string {
	if unit == 0 {
		unit = inferUnit(d)
	}

	switch unit {
	case time.Nanosecond:
		return strconv.FormatInt(d.Nanoseconds(), 10) + "ns"
	case time.Microsecond:
		usec := d / time.Microsecond
		nusec := d % time.Microsecond

		return strconv.FormatFloat(float64(usec)+float64(nusec)/1e3, 'f', 2, 64) + "µs"
	case time.Millisecond:
		msec := d / time.Millisecond
		nmsec := d % time.Millisecond

		return strconv.FormatFloat(float64(msec)+float64(nmsec)/1e6, 'f', 2, 64) + "ms"
	case time.Second:
		return strconv.FormatFloat(d.Seconds(), 'f', 2, 64) + "s"
	case time.Minute:
		return strconv.FormatFloat(d.Minutes(), 'f', 2, 64) + "m"
	default:
		panic(fmt.Errorf("invalid unit %s, should have matched one of the pre-defined unit", unit))
	}
}

func inferUnit(d time.Duration) time.Duration {
	if d < 1*time.Microsecond {
		return time.Nanosecond
	}

	if d < 1*time.Millisecond {
		return time.Microsecond
	}

	if d < 1*time.Second {
		return time.Millisecond
	}

	if d < 1*time.Minute {
		return time.Second
	}

	return time.Minute
}
