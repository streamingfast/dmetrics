package dmetrics

import "time"

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
