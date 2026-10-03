package showcase

import (
	"context"
	"encoding/json"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)

func candidate(id, source, reason string, pin int) Candidate {
	return Candidate{Item: Item{Key: id, Source: source, Reason: reason, Sponsored: reason == "sponsored", GameID: id, Tags: []string{}}, Weight: 100, Pin: pin}
}
func TestComposerDeterminismAndConstraints(t *testing.T) {
	input := Inputs{}
	for i := 0; i < 9; i++ {
		input.Candidates = append(input.Candidates, candidate("campaign:"+strconv.Itoa(i+1), "managed", []string{"editorial", "sponsored"}[i%2], 0))
	}
	for i := 0; i < 21; i++ {
		input.Candidates = append(input.Candidates, candidate("auto:"+strconv.Itoa(i+1), "automatic", []string{"upcoming", "new_release", "trending"}[i%3], 0))
	}
	want := Compose(testNow, "3", "zh", "CN", input, nil)
	if len(want.Items) != 4 {
		t.Fatal(want)
	}
	for run := 0; run < 100; run++ {
		for i, j := 0, len(input.Candidates)-1; i < j; i, j = i+1, j-1 {
			input.Candidates[i], input.Candidates[j] = input.Candidates[j], input.Candidates[i]
		}
		got := Compose(testNow, "3", "zh", "CN", input, nil)
		if !reflect.DeepEqual(want, got) {
			t.Fatal("unstable composition")
		}
		sponsored, managed := 0, 0
		seen := map[string]bool{}
		for _, item := range got.Items {
			if item.Sponsored {
				sponsored++
			}
			if item.Source == "managed" {
				managed++
			}
			if seen[item.GameID] {
				t.Fatal("duplicate Game")
			}
			seen[item.GameID] = true
		}
		if sponsored > 1 || managed > 2 {
			t.Fatal("balance violated")
		}
	}
	again := Compose(testNow.Add(time.Minute), "3", "zh", "CN", input, nil)
	if again.SnapshotID != want.SnapshotID {
		t.Fatal("hard recomposition changed identity")
	}
}
func TestComposerSmallPoolsPinsAndHints(t *testing.T) {
	for _, n := range []int{0, 1, 4, 9} {
		input := Inputs{}
		for i := 0; i < n; i++ {
			input.Candidates = append(input.Candidates, candidate(strconv.Itoa(i), "automatic", "upcoming", 0))
		}
		got := Compose(testNow, "0", "en", "US", input, nil)
		if len(got.Items) != min(n, 4) || got.Items == nil {
			t.Fatal("empty/size contract", got)
		}
	}
	for pin := 1; pin <= 4; pin++ {
		input := Inputs{Candidates: []Candidate{candidate("p", "managed", "editorial", pin)}}
		got := Compose(testNow, "0", "zh", "CN", input, nil)
		if len(got.Items) != 1 || got.Items[0].Position != pin {
			t.Fatal("pin moved")
		}
	}
	input := Inputs{Candidates: []Candidate{candidate("a", "managed", "editorial", 1), candidate("b", "managed", "editorial", 2), candidate("c", "managed", "sponsored", 3), candidate("d", "automatic", "trending", 0), candidate("e", "managed", "sponsored", 4)}}
	got := Compose(testNow, "0", "zh", "CN", input, nil)
	if len(got.Items) != 4 || got.Items[3].GameID != "d" {
		t.Fatal("pinned balance exception", got)
	}
	input = Inputs{Candidates: []Candidate{candidate("same", "automatic", "upcoming", 0), candidate("other", "automatic", "upcoming", 0)}}
	hint := HomeHint{"upcoming": {"same": true}}
	got = Compose(testNow, "0", "zh", "CN", input, hint)
	if len(got.Items) != 2 {
		t.Fatal("backfill lost candidates")
	}
	// With only one slot available the preferred representative wins.
	input.Candidates = append(input.Candidates, candidate("p1", "managed", "editorial", 1), candidate("p2", "managed", "editorial", 2), candidate("p3", "managed", "editorial", 3))
	got = Compose(testNow, "0", "zh", "CN", input, hint)
	if got.Items[3].GameID != "other" {
		t.Fatal("first-screen exclusion ignored")
	}
	hint["upcoming"]["other"] = true
	got = Compose(testNow, "0", "zh", "CN", input, hint)
	if len(got.Items) != 4 {
		t.Fatal("empty preferred pool did not fall back")
	}
}
func TestComposerBucketsAndBoundaries(t *testing.T) {
	input := Inputs{}
	for i := 0; i < 30; i++ {
		input.Candidates = append(input.Candidates, candidate(strconv.Itoa(i), "automatic", []string{"upcoming", "new_release", "trending"}[i%3], 0))
	}
	hourly, daily := map[string]bool{}, map[string]bool{}
	for i := 0; i < 24; i++ {
		hourly[Compose(testNow.Add(time.Duration(i)*time.Hour), "0", "zh", "CN", input, nil).SnapshotID] = true
		daily[Compose(testNow.AddDate(0, 0, i), "0", "zh", "CN", input, nil).Items[0].Key] = true
	}
	if len(hourly) < 2 || len(daily) < 2 {
		t.Fatal("bucket cannot rotate")
	}
	for _, boundary := range []time.Time{testNow.Add(2 * time.Minute), testNow.Add(20 * time.Minute)} {
		input.Boundary = boundary
		s := Compose(testNow, "0", "zh", "CN", input, nil)
		if s.ValidUntil.After(testNow.Add(CacheTTL)) || s.ValidUntil.After(boundary) {
			t.Fatal("TTL crosses boundary")
		}
	}
	s := Compose(testNow.Add(59*time.Minute), "0", "zh", "CN", Inputs{}, nil)
	if !s.ValidUntil.Equal(testNow.Add(time.Hour)) {
		t.Fatal("TTL crosses UTC hour")
	}
}

func TestPinnedEditorialsKeepSponsoredRepresentative(t *testing.T) {
	input := Inputs{Candidates: []Candidate{candidate("p1", "managed", "editorial", 1), candidate("p2", "managed", "editorial", 2), candidate("s", "managed", "sponsored", 0), candidate("auto", "automatic", "upcoming", 0)}}
	s := Compose(testNow, "0", "zh", "CN", input, nil)
	if len(s.Items) != 4 {
		t.Fatal(s)
	}
	sponsored := 0
	for _, item := range s.Items {
		if item.Sponsored {
			sponsored++
		}
	}
	if sponsored != 1 {
		t.Fatal("pins suppressed the Sponsored representative")
	}
}
func testFacts(recent, baseline float64) []PlayerFact {
	out := []PlayerFact{}
	for age := 0; age < 10; age++ {
		avg := baseline
		if age < 3 {
			avg = recent
		}
		expected := int32(10)
		out = append(out, PlayerFact{testNow.Truncate(24*time.Hour).AddDate(0, 0, -age), avg, 10, &expected})
	}
	return out
}
func TestTrendingFrozenV1(t *testing.T) {
	through := testNow.Truncate(24 * time.Hour)
	for _, tc := range []struct {
		name string
		r, b float64
		want bool
	}{{"tiny", 4, 1, false}, {"baseline", 20, 4, false}, {"recent", 19, 5, false}, {"delta", 20, 14, false}, {"ratio", 100, 80, false}, {"reliable", 700, 100, true}, {"threshold", 20, 10, true}} {
		t.Run(tc.name, func(t *testing.T) {
			tr := Trending(testFacts(tc.r, tc.b), through)
			if tr.Eligible != tc.want {
				t.Fatal(tr)
			}
		})
	}
	for _, which := range []string{"recent", "baseline", "coverage"} {
		f := testFacts(700, 100)
		switch which {
		case "recent":
			f[0].Successful = 0
			f[1].Successful = 0
		case "baseline":
			for i := 3; i < 7; i++ {
				f[i].Successful = 0
			}
		case "coverage":
			for i := range f {
				f[i].Successful = 4
			}
		}
		if Trending(f, through).Eligible {
			t.Fatal("poor evidence accepted", which)
		}
	}
	f := testFacts(700, 100)
	for i := range f {
		f[i].Expected = nil
	}
	if !Trending(f, through).Eligible {
		t.Fatal("legacy observed days rejected")
	}
	if Trending(f[:3], through).Eligible {
		t.Fatal("new tracking identity borrowed old baseline")
	}
	f = testFacts(700, 100)
	f[0].Average = 100
	f[0].Successful = 1
	tr := Trending(f, through)
	if tr.Recent.Average != (100.0+7000+7000)/21 {
		t.Fatal("average not sample weighted", tr)
	}
}
func TestAutomaticLocaleAndEligibility(t *testing.T) {
	sfw := true
	availability, precision := "upcoming", "day"
	g := gamesqlc.ListShowcaseGamesRow{ID: 1, Appid: 100, ShowcaseEligible: true, Sfw: &sfw, Title: "Game", Summary: "Summary", ArtworkUrl: "https://shared.akamai.steamstatic.com/header.jpg", Tags: []string{"tag"}, Availability: &availability, Precision: &precision, ExactDate: pgtype.Date{Time: testNow.AddDate(0, 0, 10), Valid: true}, FirstAvailable: pgtype.Date{Time: testNow.AddDate(0, 0, -3).Truncate(24 * time.Hour), Valid: true}}
	for _, pool := range []string{"upcoming", "new_release", "trending"} {
		c, d := automaticCandidate(g, pool, testNow, Trending(testFacts(700, 100), testNow.Truncate(24*time.Hour)))
		if !d.Candidate || c.Weight <= 0 {
			t.Fatal(pool, d)
		}
		b, _ := json.Marshal(c.Item)
		if strings.Contains(string(b), "momentum") {
			t.Fatal("score leaked")
		}
	}
	for _, mutation := range []func(*gamesqlc.ListShowcaseGamesRow){func(g *gamesqlc.ListShowcaseGamesRow) { g.ShowcaseEligible = false }, func(g *gamesqlc.ListShowcaseGamesRow) { v := false; g.Sfw = &v }, func(g *gamesqlc.ListShowcaseGamesRow) { g.Summary = "" }, func(g *gamesqlc.ListShowcaseGamesRow) { g.ArtworkUrl = "" }, func(g *gamesqlc.ListShowcaseGamesRow) { g.Tags = []string{""} }} {
		copy := g
		mutation(&copy)
		_, d := automaticCandidate(copy, "upcoming", testNow, Trend{})
		if d.Candidate {
			t.Fatal("invalid candidate accepted")
		}
	}
}

type fixedReader struct {
	input Inputs
	err   error
	calls int
}

func (r *fixedReader) Read(context.Context, time.Time, string, string) (Inputs, error) {
	r.calls++
	return r.input, r.err
}
func TestWithoutRedisOrHome(t *testing.T) {
	r := &fixedReader{}
	s := New(r, nil)
	s.Now = func() time.Time { return testNow }
	snapshot, e := s.Snapshot(context.Background(), "zh", "CN")
	if e != nil || snapshot.Items == nil {
		t.Fatal(snapshot, e)
	}
	if _, e = (Signer{}).SignSnapshot(snapshot, "zh", "CN"); e != nil {
		t.Fatal("empty requires signing")
	}
}
