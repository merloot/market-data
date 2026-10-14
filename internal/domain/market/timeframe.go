package market

import (
	"fmt"
	"time"
)

type Timeframe string

const (
	Timeframe1h          Timeframe = "1h"
	Timeframe8h          Timeframe = "8h"
	TimeframeDay         Timeframe = "24h"
	TimeframeWeek        Timeframe = "7d"
	TimeframeTwoWeeks    Timeframe = "14d"
	TimeframeMonth       Timeframe = "30d"
	TimeframeThreeMonths Timeframe = "90d"
	TimeframeSixMonths   Timeframe = "180d"
	TimeframeYear        Timeframe = "365d"
	TimeframeCustom      Timeframe = "custom"
)

func ResolvePrevTimestamp(p Period) (time.Time, error) {
	now := time.Now().UTC()
	if p.FromDate != nil {
		return *p.FromDate, nil
	}

	switch p.Timeframe {
	case Timeframe1h:
		return now.Add(-1 * time.Hour), nil
	case TimeframeDay:
		return now.Add(-24 * time.Hour), nil
	case TimeframeWeek:
		return now.Add(-7 * 24 * time.Hour), nil
	case TimeframeMonth:
		return now.Add(-30 * 24 * time.Hour), nil
	case TimeframeCustom:
		return time.Time{}, fmt.Errorf("Custom requires FromDate")
	default:
		return time.Time{}, fmt.Errorf("Unknown timeframe: %q", p.Timeframe)
	}
}

type PeriodProvider interface {
	GetPeriod() Period
}

type Period struct {
	Timeframe Timeframe
	FromDate  *time.Time
}

func (t Timeframe) Days() int {
	switch t {
	case Timeframe1h:
		return 0
	case TimeframeDay:
		return 1
	case TimeframeWeek:
		return 7
	case TimeframeMonth:
		return 30
	default:
		return 0
	}
}

func (t Timeframe) Duration() (time.Duration, error) {
	switch t {
	case Timeframe8h:
		return 8 * time.Hour, nil
	case TimeframeDay:
		return 24 * time.Hour, nil
	case TimeframeWeek:
		return 7 * 24 * time.Hour, nil
	case TimeframeTwoWeeks:
		return 14 * 24 * time.Hour, nil
	case TimeframeMonth:
		return 30 * 24 * time.Hour, nil
	case TimeframeThreeMonths:
		return 90 * 24 * time.Hour, nil
	case TimeframeSixMonths:
		return 180 * 24 * time.Hour, nil
	case TimeframeYear:
		return 365 * 24 * time.Hour, nil
	case TimeframeCustom:
		return 0, fmt.Errorf("custom timeframe requires explicit dates")
	default:
		return 0, fmt.Errorf("unknown timeframe: %q", t)
	}
}

type Bucket time.Duration

const (
	Bucket5m Bucket = Bucket(5 * time.Minute)
	Bucket1h Bucket = Bucket(time.Hour)
	Bucket3h Bucket = Bucket(3 * time.Hour)
	Bucket1d Bucket = Bucket(24 * time.Hour)
)

func (b Bucket) Strings() string {
	switch b {
	case Bucket5m:
		return "5 minutes"
	case Bucket1h:
		return "1 hour"
	case Bucket3h:
		return "3 hours"
	case Bucket1d:
		return "1 day"
	default:
		return fmt.Sprintf("%d seconds", int(time.Duration(b).Seconds()))
	}
}

func CalculateBucket(timeframe Timeframe, from, to time.Time) Bucket {
	if timeframe == TimeframeCustom {
		return calculateCustomBucket(from, to)
	}
	switch timeframe {
	case Timeframe8h, TimeframeDay:
		return Bucket5m
	case TimeframeWeek, TimeframeTwoWeeks, TimeframeMonth:
		return Bucket1h
	case TimeframeThreeMonths:
		return Bucket3h
	case TimeframeSixMonths, TimeframeYear:
		return Bucket1d
	default:
		return Bucket1h
	}
}

// TODO
func calculateCustomBucket(from, to time.Time) Bucket {
	diff := to.Sub(from)
	hours := diff.Hours()

	switch {
	case hours <= 24:
		return Bucket5m
	case diff.Hours()/24 <= 30:
		return Bucket1h
	case diff.Hours()/24 <= 90:
		return Bucket3h
	default:
		return Bucket1d
	}
}

