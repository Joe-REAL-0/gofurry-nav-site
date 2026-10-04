package showcase

import (
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func diagnosticFixture(id, name string, reasons ...string) Diagnostic {
	return Diagnostic{GameID: id, Name: name, Candidate: len(reasons) == 0, ExcludedReasons: reasons, Trending: &Trend{Weight: 100}}
}

func TestDiagnosticFiltersBeforePagination(t *testing.T) {
	input := []Diagnostic{diagnosticFixture("2", "Wolf", "adult", "not_approved"), diagnosticFixture("100", "Wolf", "not_approved"), diagnosticFixture("10", "Wolf"), diagnosticFixture("20", "Fox")}
	q, err := ParseDiagnosticQuery(url.Values{"pool": {"trending"}, "status": {"pending_approval"}, "keyword": {" wOLF "}, "page_size": {"1"}})
	if err != nil {
		t.Fatal(err)
	}
	page := QueryDiagnostics(input, q)
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].GameID != "100" || page.Items[0].Status != "pending_approval" || page.Counts != (DiagnosticCounts{Eligible: 1, PendingApproval: 1, Blocked: 1, All: 3}) {
		t.Fatalf("filtered page: %+v", page)
	}
	q.Status, q.Keyword = "all", "#100"
	if p := QueryDiagnostics(input, q); p.Total != 1 || p.Items[0].GameID != "100" {
		t.Fatal(p)
	}
	q.Keyword, q.ExcludedReason = "", "adult"
	if p := QueryDiagnostics(input, q); p.Total != 1 || p.Items[0].GameID != "2" || p.Counts.Blocked != 1 {
		t.Fatal(p)
	}
	q.Page = 1000000
	if p := QueryDiagnostics(input, q); p.Total != 1 || p.Items == nil || len(p.Items) != 0 {
		t.Fatal("out of range must keep filtered total and empty array", p)
	}
	q.Keyword = "不存在"
	if p := QueryDiagnostics(input, q); p.Total != 0 || p.Items == nil || p.Counts.All != 0 {
		t.Fatal(p)
	}
	if input[0].GameID != "2" || input[0].Status != "" {
		t.Fatal("input mutated")
	}
}

func TestDiagnosticSortingAndLegacyDefaults(t *testing.T) {
	day := func(value string) *string { return &value }
	input := []Diagnostic{diagnosticFixture("100", "B"), diagnosticFixture("10", "A"), diagnosticFixture("2", "A"), diagnosticFixture("20", "pending", "not_approved")}
	input[0].Release = &Release{ExactDate: day("2026-11-01")}
	input[1].Release = &Release{WindowStart: day("2026-10-10")}
	input[0].FirstAvailable, input[1].FirstAvailable = day("2026-10-03"), day("2026-10-01")
	input[1].Trending.Weight = 200
	input[3].Trending.Weight = 10000 // pending must not displace eligible in pool order
	for _, tc := range []struct {
		pool, sort string
		want       []string
	}{
		{"upcoming", "pool", []string{"10", "100", "2", "20"}},
		{"new_release", "pool", []string{"100", "10", "2", "20"}},
		{"trending", "pool", []string{"10", "2", "100", "20"}},
		{"trending", "name", []string{"2", "10", "100", "20"}},
		{"trending", "", []string{"2", "10", "20", "100"}},
	} {
		t.Run(tc.pool+tc.sort, func(t *testing.T) {
			q, err := ParseDiagnosticQuery(url.Values{"pool": {tc.pool}, "sort": {tc.sort}})
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				page := QueryDiagnostics(input, q)
				ids := []string{}
				for _, d := range page.Items {
					ids = append(ids, d.GameID)
				}
				if !reflect.DeepEqual(ids, tc.want) {
					t.Fatal(ids, tc.want)
				}
				slices.Reverse(input)
			}
			q.Page, q.Size = 2, 1
			if p := QueryDiagnostics(input, q); p.Items[0].GameID != tc.want[1] || p.Total != 4 {
				t.Fatal(p)
			}
		})
	}
}

func TestDiagnosticQueryValidation(t *testing.T) {
	for key, values := range map[string][]string{
		"pool": {"", "editorial"}, "status": {"ready", "all OR true"}, "sort": {"random", "weight"},
		"excluded_reason": {"unknown", "new_release_requirements"}, "keyword": {strings.Repeat("猫", 101)},
		"page_num": {"0", "-1", "1.5", "1000001", "overflow"}, "page_size": {"0", "101", "bad"},
	} {
		for _, value := range values {
			v := url.Values{"pool": {"trending"}}
			v.Set(key, value)
			if _, err := ParseDiagnosticQuery(v); err == nil {
				t.Fatalf("accepted %s=%s", key, value)
			}
		}
	}
}

func TestDiagnosticPoolExplanationsDoNotChangeEligibility(t *testing.T) {
	today := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	expired := "2026-09-01"
	d := Diagnostic{ExcludedReasons: []string{"new_release_requirements"}, FirstAvailable: &expired}
	if got := diagnosticPoolFailures(d, "new_release", today); !slices.Equal(got, []string{"first_available_expired"}) {
		t.Fatal(got)
	}
	d.FirstAvailable = nil
	if got := diagnosticPoolFailures(d, "new_release", today); !slices.Equal(got, []string{"first_available_missing"}) {
		t.Fatal(got)
	}
	future := "2026-10-06"
	d.FirstAvailable = &future
	if got := diagnosticPoolFailures(d, "new_release", today); !slices.Equal(got, []string{"first_available_future"}) {
		t.Fatal(got)
	}
	d = Diagnostic{ExcludedReasons: []string{"trending_requirements"}, Trending: &Trend{Recent: TrendWindow{ObservedDays: 3, Average: 32}, Baseline: TrendWindow{ObservedDays: 7, Average: 31}}}
	if got := diagnosticPoolFailures(d, "trending", today); !slices.Equal(got, []string{"player_delta_low", "player_ratio_low"}) {
		t.Fatal(got)
	}
	d = Diagnostic{ExcludedReasons: []string{"not_approved"}, Trending: &Trend{Eligible: true}}
	if got := diagnosticPoolFailures(d, "trending", today); len(got) != 0 || diagnosticStatus(d) != "pending_approval" {
		t.Fatal(got)
	}
	d = Diagnostic{ExcludedReasons: []string{"trending_requirements"}, Trending: &Trend{}}
	if got := diagnosticPoolFailures(d, "trending", today); !slices.Equal(got, []string{"player_facts_missing"}) {
		t.Fatal("missing facts must not be described as zero audience", got)
	}
}
