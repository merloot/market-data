package storage

import (
	"fmt"
	"time"
)

type PeriodProvider interface {
	GetPeriod() Period
}

type Period struct {
	Timeframe Timeframe
	FromDate  *time.Time
}

type MarketDataFilter struct {
	Currencies []string
	Period
}

type MarketDataRow struct {
	Currency              string
	LogoURL               *string
	Price                 string
	MarketCap             string
	CirculatingSupply     string
	TotalSupply           string
	PriceDiff             string
	MarketCapDiff         string
	CirculatingSupplyDiff string
}

type Timeframe string

const (
	Timeframe1h     Timeframe = "1h"
	Timeframe24h    Timeframe = "24h"
	Timeframe7d     Timeframe = "7d"
	Timeframe30d    Timeframe = "30d"
	TimeframeCustom Timeframe = "custom"
)

func (t Timeframe) Days() int {
	switch t {
	case Timeframe1h:
		return 0
	case Timeframe24h:
		return 1
	case Timeframe7d:
		return 7
	case Timeframe30d:
		return 30
	default:
		return 0
	}
}

func ResolvePrevTimestamp(p Period) (time.Time, error) {
	now := time.Now().UTC()
	if p.FromDate != nil {
		return *p.FromDate, nil
	}

	switch p.Timeframe {
	case Timeframe1h:
		return now.Add(-1 *time.Hour), nil
	case Timeframe24h:
		return now.Add(-24 * time.Hour), nil
	case Timeframe7d:
		return now.Add(-7 * 24 * time.Hour), nil
	case Timeframe30d:
		return now.Add(- 30 * 24 * time.Hour), nil
	case TimeframeCustom:
			return time.Time{}, fmt.Errorf("Custom requires FromDate")
	default:
			return time.Time{}, fmt.Errorf("Unknown timeframe: %q", p.Timeframe)
	}
}
