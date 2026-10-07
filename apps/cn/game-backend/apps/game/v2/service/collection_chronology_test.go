package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

func collectionRelease(availability, precision, start, end string) *v2models.GameV2ReleaseState {
	r := &v2models.GameV2ReleaseState{Availability: availability, Precision: precision}
	if start != "" {
		r.WindowStart = &start
	}
	if end != "" {
		r.WindowEnd = &end
	}
	return r
}

func TestCollectionChronologyMatrix(t *testing.T) {
	asOf := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	first := &v2models.GameV2FirstAvailable{Precision: "year", WindowStart: "2019-01-01", WindowEnd: "2019-12-31", Source: "steam_backfill", Inferred: true}
	for _, tc := range []struct {
		name                     string
		first                    *v2models.GameV2FirstAvailable
		release                  *v2models.GameV2ReleaseState
		phase, precision, source string
		inferred                 bool
	}{
		{"first available", first, nil, "released", "year", "first_available", true},
		{"observed first available", &v2models.GameV2FirstAvailable{Precision: "day", WindowStart: "2019-01-01", WindowEnd: "2019-01-01", Source: "observed_transition"}, nil, "released", "day", "first_available", false},
		{"first wins over upcoming", first, collectionRelease("upcoming", "day", "2027-01-01", "2027-01-01"), "released", "year", "first_available", true},
		{"available day", nil, collectionRelease("available", "day", "2026-10-05", "2026-10-05"), "released", "day", "release", false},
		{"available month", nil, collectionRelease("available", "month", "2026-09-01", "2026-09-30"), "released", "month", "release", false},
		{"available quarter", nil, collectionRelease("available", "quarter", "2026-07-01", "2026-09-30"), "released", "quarter", "release", false},
		{"available year", nil, collectionRelease("available", "year", "2025-01-01", "2025-12-31"), "released", "year", "release", false},
		{"available untimed", nil, collectionRelease("available", "none", "", ""), "released_unknown", "", "", false},
		{"available incomplete", nil, collectionRelease("available", "month", "2026-10-01", ""), "released_unknown", "", "", false},
		{"upcoming future", nil, collectionRelease("upcoming", "year", "2027-01-01", "2027-12-31"), "upcoming", "year", "release", false},
		{"upcoming current window", nil, collectionRelease("upcoming", "month", "2026-10-01", "2026-10-31"), "upcoming", "month", "release", false},
		{"end equals today", nil, collectionRelease("upcoming", "day", "2026-10-06", "2026-10-06"), "upcoming", "day", "release", false},
		{"end before today", nil, collectionRelease("upcoming", "day", "2026-10-05", "2026-10-05"), "upcoming_overdue", "day", "release", false},
		{"upcoming tba", nil, collectionRelease("upcoming", "tba", "", ""), "upcoming_tba", "", "", false},
		{"upcoming missing window", nil, collectionRelease("upcoming", "day", "", ""), "upcoming_tba", "", "", false},
		{"untrusted precision", nil, collectionRelease("upcoming", "unknown", "2026-10-01", "2026-10-01"), "upcoming_tba", "", "", false},
		{"invalid date", nil, collectionRelease("upcoming", "day", "2026-02-30", "2026-02-30"), "upcoming_tba", "", "", false},
		{"inverted window", nil, collectionRelease("upcoming", "day", "2026-10-06", "2026-10-05"), "upcoming_tba", "", "", false},
		{"unknown with date text", nil, &v2models.GameV2ReleaseState{Availability: "unknown", RawText: "2026-10-01"}, "unknown", "", "", false},
		{"unknown with window", nil, collectionRelease("unknown", "day", "2026-10-01", "2026-10-01"), "unknown", "", "", false},
		{"missing release", nil, nil, "unknown", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			phase, chronology := resolveCollectionChronology(tc.first, tc.release, asOf)
			if phase != tc.phase {
				t.Fatalf("phase=%s want %s", phase, tc.phase)
			}
			if tc.precision == "" {
				if chronology != nil {
					t.Fatalf("invented chronology: %+v", chronology)
				}
				return
			}
			if chronology == nil || chronology.Precision != tc.precision || chronology.Source != tc.source || chronology.Inferred != tc.inferred {
				t.Fatalf("chronology=%+v", chronology)
			}
			if tc.first != nil && (chronology.WindowStart != tc.first.WindowStart || chronology.WindowEnd != tc.first.WindowEnd) {
				t.Fatal("first available window changed")
			}
			if tc.first == nil && (chronology.WindowStart != *tc.release.WindowStart || chronology.WindowEnd != *tc.release.WindowEnd) {
				t.Fatal("release window changed")
			}
			body, _ := json.Marshal(chronology)
			for _, forbidden := range []string{"exact_date", "display_date", "steam_backfill", "legacy_manual", "observed_transition"} {
				if strings.Contains(string(body), forbidden) {
					t.Fatal("internal/synthetic date metadata leaked", string(body))
				}
			}
		})
	}
	// Local midnight is not a UTC date boundary.
	local := time.Date(2026, 10, 7, 1, 0, 0, 0, time.FixedZone("CST", 8*3600))
	phase, _ := resolveCollectionChronology(nil, collectionRelease("upcoming", "day", "2026-10-06", "2026-10-06"), local)
	if phase != "upcoming" {
		t.Fatal("asOf was not interpreted in UTC")
	}
}

func TestCollectionTimelineOrderingAndPreview(t *testing.T) {
	window := func(start, end string) *v2models.CollectionChronology {
		return &v2models.CollectionChronology{WindowStart: start, WindowEnd: end}
	}
	items := []v2models.CollectionTimelineItem{
		{GameID: "20", Phase: "unknown"}, {GameID: "19", Phase: "upcoming_tba"},
		{GameID: "18", Phase: "upcoming", Chronology: window("2027-01-01", "2027-12-31")},
		{GameID: "17", Phase: "upcoming_overdue", Chronology: window("2026-01-01", "2026-01-01")},
		{GameID: "16", Phase: "released_unknown"}, {GameID: "10", Phase: "released", Chronology: window("2025-01-01", "2025-12-31")},
		{GameID: "2", Phase: "released", Chronology: window("2025-01-01", "2025-12-31")},
		{GameID: "3", Phase: "released", Chronology: window("2025-01-01", "2025-01-31")},
		{GameID: "99", Phase: "released", Chronology: window("2024-01-01", "2024-12-31")},
		{GameID: "1", Phase: "unknown"}, {GameID: "1", Phase: "upcoming_tba"}, {GameID: "1", Phase: "released_unknown"},
	}
	want := []string{"99", "3", "2", "10", "1", "16", "17", "18", "1", "19", "1", "20"}
	for rotation := 0; rotation < len(items); rotation++ {
		shuffled := append(append([]v2models.CollectionTimelineItem{}, items[rotation:]...), items[:rotation]...)
		sortCollectionTimeline(shuffled)
		var got []string
		for _, item := range shuffled {
			got = append(got, item.GameID)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ordering=%v", got)
		}
	}
	for _, tc := range []struct {
		count int
		want  []string
	}{{0, []string{}}, {1, []string{"20"}}, {2, []string{"20", "19"}}, {3, []string{"20", "19", "18"}}, {4, []string{"20", "18", "17"}}, {5, []string{"20", "18", "16"}}} {
		preview := collectionPreview(items[:tc.count])
		got := []string{}
		for _, p := range preview {
			got = append(got, p.GameID)
		}
		if preview == nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("count %d preview=%v", tc.count, got)
		}
	}
}
