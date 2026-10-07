package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSearchSuggestionsCacheStates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items []string
		err   error
		state string
		calls int64
	}{
		{"ready", []string{"  furry  ", "furry", "", "兽人"}, nil, "ready", 1},
		{"empty", []string{}, nil, "empty", 1},
		{"unavailable", nil, errors.New("provider unavailable"), "unavailable", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := newMemorySuggestionCache()
			provider := &fakeSuggestionProvider{items: tc.items, err: tc.err}
			svc := newSearchService(provider, cache, fixedSearchNow)
			first := svc.GetSearchSuggestions(context.Background(), "  兽人  ")
			second := svc.GetSearchSuggestions(context.Background(), "兽人")
			if first.SchemaVersion != 2 || first.Query != "兽人" || first.State != tc.state || second.State != tc.state || first.CacheState != "miss" || first.Suggestions == nil {
				t.Fatalf("unexpected responses: %#v %#v", first, second)
			}
			if provider.calls.Load() != tc.calls {
				t.Fatalf("provider calls = %d", provider.calls.Load())
			}
			if tc.err == nil {
				if second.CacheState != "hit" || cache.ttl != 10*time.Minute || cache.sets != 1 {
					t.Fatalf("cache hit/TTL/write count: %#v %#v", second, cache)
				}
			} else if cache.sets != 0 || second.CacheState != "miss" {
				t.Fatal("unavailable was cached")
			}
			if tc.name == "ready" && strings.Join(first.Suggestions, ",") != "furry,兽人" {
				t.Fatal(first.Suggestions)
			}
		})
	}
}

func TestSearchSuggestionsUsesCacheBeforeProvider(t *testing.T) {
	cache := newMemorySuggestionCache()
	cache.Set(searchSuggestionCacheKey("兽人"), []string{"cached"}, 10*time.Minute)
	provider := &fakeSuggestionProvider{}
	svc := newSearchService(provider, cache, fixedSearchNow)
	response := svc.GetSearchSuggestions(context.Background(), "兽人")
	if response.State != "ready" || response.CacheState != "hit" || response.Suggestions[0] != "cached" || provider.calls.Load() != 0 {
		t.Fatalf("unexpected cache result: %#v", response)
	}
	response.Suggestions[0] = "modified"
	if svc.GetSearchSuggestions(context.Background(), "兽人").Suggestions[0] != "cached" {
		t.Fatal("response mutated cache")
	}
}

func TestSearchSuggestionsQueryAndNamespace(t *testing.T) {
	provider := &fakeSuggestionProvider{items: []string{"ok"}}
	svc := newSearchService(provider, newMemorySuggestionCache(), fixedSearchNow)
	response := svc.GetSearchSuggestions(context.Background(), "  "+strings.Repeat("兽", 140)+"  ")
	if len([]rune(response.Query)) != 128 || provider.lastQuery != response.Query {
		t.Fatal("query must be normalized before fetching")
	}
	key := searchSuggestionCacheKey(response.Query)
	if !strings.HasPrefix(key, "nav:v2:search:suggestions:v2:") || strings.Contains(key, "兽") {
		t.Fatal("wrong versioned query hash")
	}
	empty := svc.GetSearchSuggestions(context.Background(), "  ")
	if empty.State != "empty" || empty.Suggestions == nil || provider.calls.Load() != 1 {
		t.Fatal("blank query called provider")
	}
}

func TestSearchSuggestionsSingleflightAndCanceledWaiter(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	provider := &fakeSuggestionProvider{items: []string{"furry"}, started: started, release: release}
	svc := newSearchService(provider, newMemorySuggestionCache(), fixedSearchNow)
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan string, 1)
	go func() { first <- svc.GetSearchSuggestions(ctx, "furry").State }()
	<-started
	cancel()
	if <-first != "unavailable" {
		t.Fatal("canceled waiter did not return")
	}
	const count = 12
	var ready, done sync.WaitGroup
	ready.Add(count)
	done.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer done.Done()
			ready.Done()
			response := svc.GetSearchSuggestions(context.Background(), "furry")
			if response.State != "ready" {
				t.Errorf("waiter state: %s", response.State)
			}
		}()
	}
	ready.Wait()
	close(release)
	done.Wait()
	if provider.calls.Load() != 1 {
		t.Fatalf("upstream requests: %d", provider.calls.Load())
	}
}

func TestInvalidProxyIsUnavailableAndNotCached(t *testing.T) {
	cache := newMemorySuggestionCache()
	svc := newSearchService(newDuckDuckGoProvider("://invalid"), cache, fixedSearchNow)
	if svc.GetSearchSuggestions(context.Background(), "furry").State != "unavailable" || cache.sets != 0 {
		t.Fatal("invalid proxy must fail closed")
	}
}

func fixedSearchNow() time.Time { return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC) }

type memorySuggestionCache struct {
	mu    sync.Mutex
	items map[string][]string
	ttl   time.Duration
	sets  int
}

func newMemorySuggestionCache() *memorySuggestionCache {
	return &memorySuggestionCache{items: map[string][]string{}}
}
func (c *memorySuggestionCache) Get(key string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	items, ok := c.items[key]
	return copySuggestions(items), ok
}
func (c *memorySuggestionCache) Set(key string, items []string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = copySuggestions(items)
	c.ttl = ttl
	c.sets++
}

type fakeSuggestionProvider struct {
	items     []string
	err       error
	calls     atomic.Int64
	lastQuery string
	started   chan struct{}
	release   chan struct{}
}

func (p *fakeSuggestionProvider) Fetch(ctx context.Context, query string) ([]string, error) {
	p.calls.Add(1)
	p.lastQuery = query
	if p.started != nil {
		close(p.started)
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return p.items, p.err
}
