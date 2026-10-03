package showcase

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/redis/go-redis/v9"
)

func localRedis(t *testing.T, db int) *redis.Client {
	t.Helper()
	addr := os.Getenv("GOFURRY_SHOWCASE_REDIS_ADDR")
	if addr == "" {
		t.Skip("set GOFURRY_SHOWCASE_REDIS_ADDR to dedicated disposable localhost Redis")
	}
	if !strings.HasPrefix(addr, "127.0.0.1:") {
		t.Fatal("Showcase test Redis must be local and disposable")
	}
	c := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	t.Cleanup(func() { c.FlushDB(context.Background()); c.Close() })
	if e := c.FlushDB(context.Background()).Err(); e != nil {
		t.Fatal(e)
	}
	return c
}

type statMemory struct {
	stats   map[string]gamesqlc.UpsertShowcaseDailyStatParams
	quality map[string]gamesqlc.UpsertShowcaseQualityParams
}

func (w *statMemory) UpsertShowcaseDailyStat(_ context.Context, p gamesqlc.UpsertShowcaseDailyStatParams) error {
	w.stats[p.StatDate.Time.Format(time.DateOnly)+p.SubjectKey] = p
	return nil
}
func (w *statMemory) UpsertShowcaseQuality(_ context.Context, p gamesqlc.UpsertShowcaseQualityParams) error {
	w.quality[p.StatDate.Time.Format(time.DateOnly)] = p
	return nil
}
func TestShowcaseRedisQualificationAndAggregation(t *testing.T) {
	c := localRedis(t, 12)
	ctx := context.Background()
	clock := testNow
	writer := &statMemory{map[string]gamesqlc.UpsertShowcaseDailyStatParams{}, map[string]gamesqlc.UpsertShowcaseQualityParams{}}
	a := NewAnalytics(c, writer, testSigner(), []string{"https://example.test"})
	a.Now = func() time.Time { return clock }
	snap := signedTestSnapshot(t)
	token := snap.Items[0].TrackingToken
	session := "11111111-1111-4111-8111-111111111111"
	send := func(event, source, session string) {
		a.Submit(ctx, Event{token, session, event, source}, "https://example.test", "Mozilla/5.0 Chrome/120.0 Safari/537.36", "192.0.2.1")
	}
	send("impression", "", session)
	send("impression", "", session)
	send("click", "primary", session)
	send("click", "artwork", session)
	send("click", "title", "22222222-2222-4222-8222-222222222222")
	send("click", "secondary", "33333333-3333-4333-8333-333333333333")
	if e := a.Aggregate(ctx); e != nil {
		t.Fatal(e)
	}
	p := writer.stats["2026-10-03campaign:1"]
	if p.ValidImpressions != 2 || p.QualifiedClicks != 2 || p.ClickPrimary != 1 || p.ClickTitle != 1 || p.SessionEstimate != 2 {
		t.Fatal("qualification invariant", p)
	}
	quality := writer.quality["2026-10-03"]
	if quality.DuplicateImpressions != 1 || quality.DuplicateClicks != 1 || quality.InvalidTokenEvents != 1 {
		t.Fatal("quality", quality)
	}
	before := p
	for i := 0; i < 3; i++ {
		if e := a.Aggregate(ctx); e != nil {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(writer.stats["2026-10-03campaign:1"], before) {
		t.Fatal("additive replay")
	}
	keys, e := c.Keys(ctx, "game:v2:showcase:*").Result()
	if e != nil {
		t.Fatal(e)
	}
	for _, k := range keys {
		if strings.Contains(k, session) || strings.Contains(k, "192.0.2.1") {
			t.Fatal("raw identity in Redis key")
		}
		ttl := c.TTL(ctx, k).Val()
		if ttl <= 0 {
			t.Fatal("unbounded key retention", k)
		}
	}
	clock = clock.AddDate(0, 0, 2)
	writer.stats = map[string]gamesqlc.UpsertShowcaseDailyStatParams{}
	if e := a.Aggregate(ctx); e != nil {
		t.Fatal(e)
	}
	if writer.stats["2026-10-03campaign:1"].QualifiedClicks != 2 {
		t.Fatal("prior-two-day recovery lost counters")
	}
	// Redis failure remains a discarded side effect; Submit has no navigation error.
	_ = c.Close()
	a.Submit(ctx, Event{token, session, "impression", ""}, "https://example.test", "Mozilla/5.0 Chrome/120.0", "192.0.2.1")
}
func TestShowcaseRedisFiltersRatesAndMidnight(t *testing.T) {
	c := localRedis(t, 12)
	ctx := context.Background()
	clock := testNow
	a := NewAnalytics(c, nil, testSigner(), []string{"https://example.test"})
	a.Now = func() time.Time { return clock }
	token := signedTestSnapshot(t).Items[0].TrackingToken
	session := "11111111-1111-4111-8111-111111111111"
	e := Event{token, session, "impression", ""}
	a.Submit(ctx, e, "", "Mozilla/5.0 Chrome/120.0", "192.0.2.1")
	a.Submit(ctx, e, "https://example.test", "curl/8.0", "192.0.2.1")
	for i := 0; i < 31; i++ {
		a.Submit(ctx, e, "https://example.test", "Mozilla/5.0 Chrome/120.0", "192.0.2.1")
	}
	for i := 0; i < 601; i++ {
		copy := e
		copy.SessionID = fmt.Sprintf("%08x-1111-4111-8111-111111111111", i+100)
		a.Submit(ctx, copy, "https://example.test", "Mozilla/5.0 Chrome/120.0", "192.0.2.2")
	}
	q := c.HGetAll(ctx, "game:v2:showcase:quality:2026-10-03").Val()
	if q["invalid_origin_events"] != "1" || q["filtered_user_agent_events"] != "1" || q["session_rate_limited"] != "3" || q["ip_rate_limited"] != "1" {
		t.Fatal(q)
	}
	c.FlushDB(ctx)
	clock = time.Date(2026, 10, 3, 15, 59, 0, 0, time.UTC)
	snap := Compose(clock, "0", "zh", "CN", Inputs{Candidates: []Candidate{{Item: Item{Key: "campaign:1", CampaignID: "1", Source: "managed", Reason: "editorial", Tags: []string{}}, Weight: 100}}}, nil)
	signed, err := testSigner().SignSnapshot(snap, "zh", "CN")
	if err != nil {
		t.Fatal(err)
	}
	e.TrackingToken = signed.Items[0].TrackingToken
	a.Submit(ctx, e, "https://example.test", "Mozilla/5.0 Chrome/120.0", "192.0.2.1")
	clock = clock.Add(2 * time.Minute)
	e.Event, e.Source = "click", "primary"
	a.Submit(ctx, e, "https://example.test", "Mozilla/5.0 Chrome/120.0", "192.0.2.1")
	p := c.HGetAll(ctx, "game:v2:showcase:stat:2026-10-04:campaign:1").Val()
	if p["valid_impressions"] != "1" || p["qualified_clicks"] != "1" {
		t.Fatal("midnight daily invariant", p)
	}
}
func TestShowcaseRevisionCacheBoundaries(t *testing.T) {
	c := localRedis(t, 14)
	ctx := context.Background()
	clock := testNow
	r := &fixedReader{input: Inputs{Boundary: testNow.Add(2 * time.Minute)}}
	s := New(r, c)
	s.Now = func() time.Time { return clock }
	first, e := s.Snapshot(ctx, "zh", "CN")
	if e != nil {
		t.Fatal(e)
	}
	second, e := s.Snapshot(ctx, "zh", "CN")
	if e != nil || r.calls != 1 || first.SnapshotID != second.SnapshotID {
		t.Fatal("cache miss on repeat")
	}
	ttl := c.TTL(ctx, "game:v2:showcase:0:zh:CN:2026-10-03T10").Val()
	if ttl > 2*time.Minute || ttl <= 0 {
		t.Fatal("schedule TTL", ttl)
	}
	c.Incr(ctx, RevisionKey)
	third, e := s.Snapshot(ctx, "zh", "CN")
	if e != nil || r.calls != 2 || third.SnapshotID == first.SnapshotID {
		t.Fatal("revision did not invalidate")
	}
	clock = clock.Add(3 * time.Minute)
	if _, e := s.Snapshot(ctx, "zh", "CN"); e != nil || r.calls != 3 {
		t.Fatal("expired payload reused")
	}
	c.Set(ctx, "game:v2:home:zh:CN", `{"panel":{"updated_games":[{"id":"1"}],"latest_games":[{"id":"2"}],"popular_games":[{"id":"3"}]}}`, time.Hour)
	hint := s.homeHint(ctx, "zh", "CN")
	if !hint["upcoming"]["1"] || !hint["new_release"]["2"] || !hint["trending"]["3"] {
		t.Fatal(hint)
	}
}
