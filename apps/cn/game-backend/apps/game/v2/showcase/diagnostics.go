package showcase

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// DiagnosticQuery only shapes internal operator reads. Composer never uses it.
type DiagnosticQuery struct {
	Pool, Status, Keyword, ExcludedReason, Sort string
	Page, Size                                  int
}

type DiagnosticCounts struct {
	Eligible        int `json:"eligible"`
	PendingApproval int `json:"pending_approval"`
	Blocked         int `json:"blocked"`
	All             int `json:"all"`
}

type DiagnosticPage struct {
	Total  int              `json:"total"`
	Items  []Diagnostic     `json:"items"`
	Counts DiagnosticCounts `json:"counts"`
}

func ParseDiagnosticQuery(values url.Values) (DiagnosticQuery, error) {
	q := DiagnosticQuery{Pool: values.Get("pool"), Status: values.Get("status"), Keyword: strings.TrimSpace(values.Get("keyword")), ExcludedReason: values.Get("excluded_reason"), Sort: values.Get("sort"), Page: 1, Size: 20}
	if q.Status == "" {
		q.Status = "all"
	}
	if q.Sort == "" {
		q.Sort = "game_id" // preserve legacy internal clients' ordering
	}
	var err error
	if v := values.Get("page_num"); v != "" {
		q.Page, err = strconv.Atoi(v)
	}
	if err == nil {
		if v := values.Get("page_size"); v != "" {
			q.Size, err = strconv.Atoi(v)
		}
	}
	if err != nil || q.Page < 1 || q.Page > 1000000 || q.Size < 1 || q.Size > 100 || utf8.RuneCountInString(q.Keyword) > 100 ||
		!slices.Contains([]string{"upcoming", "new_release", "trending"}, q.Pool) ||
		!slices.Contains([]string{"all", "eligible", "pending_approval", "blocked"}, q.Status) ||
		!slices.Contains([]string{"pool", "name", "game_id"}, q.Sort) ||
		!slices.Contains([]string{"", "not_approved", "adult", "locale_incomplete", "steam_artwork_unavailable", q.Pool + "_requirements"}, q.ExcludedReason) {
		return q, errors.New("invalid candidate query")
	}
	return q, nil
}

func diagnosticStatus(d Diagnostic) string {
	if d.Candidate {
		return "eligible"
	}
	if len(d.ExcludedReasons) == 1 && d.ExcludedReasons[0] == "not_approved" {
		return "pending_approval"
	}
	return "blocked"
}

// Filter/count/sort the whole diagnostic set before slicing. Do not mutate the
// reader's input order (which also feeds the independently deterministic Composer).
func QueryDiagnostics(input []Diagnostic, q DiagnosticQuery) DiagnosticPage {
	result := DiagnosticPage{Items: []Diagnostic{}}
	rows := []Diagnostic{}
	keyword := strings.ToLower(q.Keyword)
	for _, row := range input {
		if keyword != "" && !strings.Contains(strings.ToLower(row.Name), keyword) && row.GameID != strings.TrimPrefix(keyword, "#") {
			continue
		}
		if q.ExcludedReason != "" && !slices.Contains(row.ExcludedReasons, q.ExcludedReason) {
			continue
		}
		row.Status = diagnosticStatus(row)
		result.Counts.All++
		switch row.Status {
		case "eligible":
			result.Counts.Eligible++
		case "pending_approval":
			result.Counts.PendingApproval++
		default:
			result.Counts.Blocked++
		}
		if q.Status == "all" || row.Status == q.Status {
			rows = append(rows, row)
		}
	}
	slices.SortFunc(rows, func(a, b Diagnostic) int {
		order := 0
		switch q.Sort {
		case "name":
			order = strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case "pool":
			rank := func(d Diagnostic) int {
				switch d.Status {
				case "eligible":
					return 0
				case "pending_approval":
					return 1
				default:
					return 2
				}
			}
			order = cmp.Compare(rank(a), rank(b))
			if order == 0 {
				if q.Pool == "trending" {
					weight := func(d Diagnostic) int {
						if d.Trending != nil {
							return d.Trending.Weight
						}
						return 0
					}
					order = cmp.Compare(weight(b), weight(a))
				} else {
					ad, bd := diagnosticDate(a, q.Pool), diagnosticDate(b, q.Pool)
					switch {
					case ad == "" && bd != "":
						order = 1
					case ad != "" && bd == "":
						order = -1
					case q.Pool == "new_release":
						order = strings.Compare(bd, ad)
					default:
						order = strings.Compare(ad, bd)
					}
				}
			}
		}
		if order != 0 {
			return order
		}
		ai, _ := strconv.ParseInt(a.GameID, 10, 64)
		bi, _ := strconv.ParseInt(b.GameID, 10, 64)
		return cmp.Compare(ai, bi)
	})
	result.Total = len(rows)
	start := min((q.Page-1)*q.Size, len(rows))
	result.Items = append(result.Items, rows[start:min(start+q.Size, len(rows))]...)
	return result
}

func diagnosticDate(d Diagnostic, pool string) string {
	if pool == "new_release" {
		return value(d.FirstAvailable)
	}
	r := d.Release
	if r == nil {
		return ""
	}
	if r.ExactDate != nil {
		return *r.ExactDate
	}
	if r.WindowStart != nil {
		return *r.WindowStart
	}
	if r.Year != nil {
		month := int32(1)
		if r.Month != nil {
			month = *r.Month
		} else if r.Quarter != nil {
			month = (*r.Quarter-1)*3 + 1
		}
		return fmt.Sprintf("%04d-%02d-01", *r.Year, month)
	}
	return ""
}

// Explain the existing frozen gates; these codes never determine eligibility.
func diagnosticPoolFailures(d Diagnostic, pool string, today time.Time) []string {
	failures := []string{}
	if !slices.Contains(d.ExcludedReasons, pool+"_requirements") {
		return failures
	}
	switch pool {
	case "upcoming":
		failures = append(failures, "not_upcoming")
	case "new_release":
		if d.FirstAvailable == nil {
			failures = append(failures, "first_available_missing")
		} else if *d.FirstAvailable > today.Format(time.DateOnly) {
			failures = append(failures, "first_available_future")
		} else {
			failures = append(failures, "first_available_expired")
		}
	case "trending":
		if d.Trending == nil || (d.Trending.Recent.ObservedDays == 0 && d.Trending.Baseline.ObservedDays == 0) {
			return []string{"player_facts_missing"}
		}
		t := d.Trending
		checks := []struct {
			failed bool
			code   string
		}{
			{t.Recent.ObservedDays < MinRecentObservedDays, "recent_days_insufficient"},
			{t.Baseline.ObservedDays < MinBaselineObservedDays, "baseline_days_insufficient"},
			{t.Recent.Coverage != nil && *t.Recent.Coverage < MinCoverage, "recent_coverage_insufficient"},
			{t.Baseline.Coverage != nil && *t.Baseline.Coverage < MinCoverage, "baseline_coverage_insufficient"},
			{t.Baseline.Average < MinBaselineAverage, "baseline_players_low"},
			{t.Recent.Average < MinRecentAverage, "recent_players_low"},
			{t.Recent.Average-t.Baseline.Average < MinDelta, "player_delta_low"},
			{t.Baseline.Average > 0 && t.Recent.Average/t.Baseline.Average < MinRatio, "player_ratio_low"},
		}
		for _, check := range checks {
			if check.failed {
				failures = append(failures, check.code)
			}
		}
	}
	return failures
}
