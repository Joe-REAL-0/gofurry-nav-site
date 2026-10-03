package showcase

import (
	"math"
	"time"
)

const (
	RecentDays              = 3
	BaselineDays            = 7
	MinRecentObservedDays   = 2
	MinBaselineObservedDays = 4
	MinCoverage             = 0.5
	MinBaselineAverage      = 5.0
	MinRecentAverage        = 20.0
	MinDelta                = 10.0
	MinRatio                = 1.5
	MomentumRatioCap        = 3.0
	MaxSelectionWeight      = 10000
)

type PlayerFact struct {
	Date       time.Time
	Average    float64
	Successful int32
	Expected   *int32
}
type TrendWindow struct {
	ObservedDays int      `json:"observed_days"`
	Average      float64  `json:"average"`
	Coverage     *float64 `json:"coverage"`
}
type Trend struct {
	Recent   TrendWindow `json:"recent"`
	Baseline TrendWindow `json:"baseline"`
	Momentum float64     `json:"momentum"`
	Eligible bool        `json:"eligible"`
	Weight   int         `json:"weight"`
}

func Trending(facts []PlayerFact, through time.Time) Trend {
	type window struct {
		days                    int
		sum                     float64
		samples                 int64
		expected, ledgerSuccess int64
		ledger                  bool
	}
	var windows [2]window
	seen := map[string]bool{}
	for _, f := range facts {
		day := f.Date.UTC().Format(time.DateOnly)
		if seen[day] {
			continue
		}
		seen[day] = true
		age := int(through.UTC().Sub(f.Date.UTC()) / (24 * time.Hour))
		if age < 0 || age >= RecentDays+BaselineDays {
			continue
		}
		index := 0
		if age >= RecentDays {
			index = 1
		}
		w := &windows[index]
		if f.Expected != nil {
			w.ledger = true
			w.expected += int64(*f.Expected)
			w.ledgerSuccess += int64(f.Successful)
		}
		if f.Successful > 0 && !math.IsNaN(f.Average) && !math.IsInf(f.Average, 0) && f.Average >= 0 {
			w.days++
			w.samples += int64(f.Successful)
			w.sum += f.Average * float64(f.Successful)
		}
	}
	convert := func(w window) TrendWindow {
		v := TrendWindow{ObservedDays: w.days}
		if w.samples > 0 {
			v.Average = w.sum / float64(w.samples)
		}
		if w.ledger {
			coverage := 0.0
			if w.expected > 0 {
				coverage = float64(w.ledgerSuccess) / float64(w.expected)
			}
			v.Coverage = &coverage
		}
		return v
	}
	t := Trend{Recent: convert(windows[0]), Baseline: convert(windows[1])}
	if t.Recent.ObservedDays < MinRecentObservedDays || t.Baseline.ObservedDays < MinBaselineObservedDays {
		return t
	}
	if t.Recent.Coverage != nil && *t.Recent.Coverage < MinCoverage || t.Baseline.Coverage != nil && *t.Baseline.Coverage < MinCoverage {
		return t
	}
	r, b := t.Recent.Average, t.Baseline.Average
	if b < MinBaselineAverage || r < MinRecentAverage || r-b < MinDelta || r/b < MinRatio {
		return t
	}
	t.Momentum = (r - b) * math.Min(r/b, MomentumRatioCap)
	t.Weight = int(math.Min(MaxSelectionWeight, math.Max(1, math.Round(t.Momentum))))
	t.Eligible = true
	return t
}
