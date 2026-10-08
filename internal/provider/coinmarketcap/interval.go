package coinmarketcap

import "time"

const (
	dayInMs   = int64(24 * time.Hour / time.Millisecond)
	monthInMs = dayInMs * 30
)

func intervalForRange(from, to time.Time) string {
	diff := to.UnixMilli() - from.UnixMilli()

	switch {
	case diff > 3*monthInMs:
		return "1d"
	case diff > 2*dayInMs:
		return "1h"
	default:
		return "5m"
	}
}
