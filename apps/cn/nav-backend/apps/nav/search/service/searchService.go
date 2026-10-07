package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/models"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	searchSuggestionMaxQueryLen = 128
	searchSuggestionCacheTTL    = 10 * time.Minute
)

type suggestionProvider interface {
	Fetch(ctx context.Context, query string) ([]string, error)
}

type suggestionCache interface {
	Get(key string) ([]string, bool)
	Set(key string, suggestions []string, ttl time.Duration)
}

type redisSuggestionCache struct{}

func (redisSuggestionCache) Get(key string) ([]string, bool) {
	raw, err := cs.GetString(key)
	if err != nil || raw == "" {
		return nil, false
	}
	var suggestions []string
	if jsonErr := sonic.Unmarshal([]byte(raw), &suggestions); jsonErr != nil {
		log.Warn("search suggestion cache unmarshal error:", jsonErr)
		return nil, false
	}
	return copySuggestions(suggestions), true
}

func (redisSuggestionCache) Set(key string, suggestions []string, ttl time.Duration) {
	payload, err := sonic.Marshal(suggestions)
	if err != nil {
		return
	}
	_ = cs.SetExpire(key, string(payload), ttl)
}

type searchService struct {
	provider suggestionProvider
	cache    suggestionCache
	group    singleflight.Group
	now      func() time.Time
}

var (
	searchSingleton = &searchService{}
	searchMu        sync.Mutex
)

func GetSearchService() *searchService {
	searchMu.Lock()
	defer searchMu.Unlock()
	if searchSingleton.provider == nil {
		searchSingleton.provider = newDuckDuckGoProvider(env.GetServerConfig().Proxy.Url)
	}
	if searchSingleton.cache == nil {
		searchSingleton.cache = redisSuggestionCache{}
	}
	if searchSingleton.now == nil {
		searchSingleton.now = time.Now
	}
	return searchSingleton
}

func newSearchService(provider suggestionProvider, cache suggestionCache, now func() time.Time) *searchService {
	return &searchService{provider: provider, cache: cache, now: now}
}

func New(proxyURL string) *searchService {
	return newSearchService(newDuckDuckGoProvider(proxyURL), redisSuggestionCache{}, time.Now)
}

func (svc *searchService) GetSearchSuggestions(ctx context.Context, query string) models.SearchSuggestionsResponse {
	query = normalizeSuggestionQuery(query)
	response := models.SearchSuggestionsResponse{
		SchemaVersion: models.SearchSuggestionsSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.SearchSuggestionsStateEmpty,
		Query:         query,
		Suggestions:   []string{},
		CacheState:    models.SearchSuggestionsCacheMiss,
	}

	if query == "" {
		return response
	}

	cacheKey := searchSuggestionCacheKey(query)
	if cached, ok := svc.cacheStore().Get(cacheKey); ok {
		response.CacheState = models.SearchSuggestionsCacheHit
		response.Suggestions = cached
		if len(cached) > 0 {
			response.State = models.SearchSuggestionsStateReady
		}
		return response
	}

	pending := svc.group.DoChan(cacheKey, func() (any, error) {
		if cached, ok := svc.cacheStore().Get(cacheKey); ok {
			return cachedSearchSuggestions{items: cached, hit: true}, nil
		}
		// One bounded upstream request survives cancellation of an individual waiter.
		fetchCtx, cancel := context.WithTimeout(context.Background(), searchSuggestTimeout)
		defer cancel()
		items, fetchErr := svc.provider.Fetch(fetchCtx, query)
		if fetchErr != nil {
			return cachedSearchSuggestions{items: []string{}, hit: false}, fetchErr
		}
		items = sanitizeSuggestions(items)
		svc.cacheStore().Set(cacheKey, items, searchSuggestionCacheTTL)
		return cachedSearchSuggestions{items: items, hit: false}, nil
	})
	var result singleflight.Result
	select {
	case <-ctx.Done():
		response.State = models.SearchSuggestionsStateUnavailable
		return response
	case result = <-pending:
	}
	if result.Err != nil {
		response.State = models.SearchSuggestionsStateUnavailable
		log.Warn("search suggestion provider unavailable:", result.Err)
		return response
	}
	data := result.Val.(cachedSearchSuggestions)
	response.Suggestions = copySuggestions(data.items)
	if data.hit {
		response.CacheState = models.SearchSuggestionsCacheHit
	}
	if len(response.Suggestions) > 0 {
		response.State = models.SearchSuggestionsStateReady
	}
	return response
}

func (svc *searchService) cacheStore() suggestionCache {
	if svc != nil && svc.cache != nil {
		return svc.cache
	}
	return redisSuggestionCache{}
}

func (svc *searchService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

type cachedSearchSuggestions struct {
	items []string
	hit   bool
}

func normalizeSuggestionQuery(q string) string {
	q = strings.TrimSpace(q)
	runes := []rune(q)
	if len(runes) <= searchSuggestionMaxQueryLen {
		return q
	}
	return string(runes[:searchSuggestionMaxQueryLen])
}

func searchSuggestionCacheKey(query string) string {
	return "nav:v2:search:suggestions:v2:" + util.CreateMD5(query)
}

func sanitizeSuggestions(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func copySuggestions(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	copied := make([]string, len(items))
	copy(copied, items)
	return copied
}

type redisSuggestionRateLimiter struct {
	client *redis.Client
	now    func() time.Time
}

func NewRedisSuggestionRateLimiter() *redisSuggestionRateLimiter {
	return &redisSuggestionRateLimiter{now: time.Now}
}

func (limiter *redisSuggestionRateLimiter) Allow(ip string) (bool, int64) {
	client := limiter.client
	if client == nil {
		client = cs.GetRedisService()
	}
	if client == nil {
		return true, 0
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now
	if limiter != nil && limiter.now != nil {
		now = limiter.now
	}
	window := now().Unix() / int64(SearchSuggestionRateWindow/time.Second)
	key := "nav:v2:search:suggestions:rate:" + util.CreateMD5(ip) + ":" + fmt.Sprint(window)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		log.Warn("search suggestion rate limiter incr error:", err)
		return true, 0
	}
	if count == 1 {
		_ = client.Expire(ctx, key, SearchSuggestionRateWindow+time.Minute).Err()
	}
	if count > SearchSuggestionRateLimit {
		return false, int64(SearchSuggestionRateWindow / time.Second)
	}
	return true, 0
}

const (
	SearchSuggestionRateLimit  = int64(300)
	SearchSuggestionRateWindow = 10 * time.Minute
)
