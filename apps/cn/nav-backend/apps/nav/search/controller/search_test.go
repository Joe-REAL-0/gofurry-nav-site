package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
)

func TestGetSearchSuggestionsReturnsV2Envelope(t *testing.T) {
	for _, params := range []string{"q=furry", "q=furry&engine=bing", "q=furry&engine=unsupported"} {
		t.Run(params, func(t *testing.T) { testSuggestionEnvelope(t, params) })
	}
}
func testSuggestionEnvelope(t *testing.T, params string) {
	now := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	reader := &fakeSuggestionsReader{response: models.SearchSuggestionsResponse{
		SchemaVersion: models.SearchSuggestionsSchemaVersion,
		GeneratedAt:   now,
		State:         models.SearchSuggestionsStateReady,
		Query:         "furry",
		Suggestions:   []string{"furry game"},
		CacheState:    models.SearchSuggestionsCacheMiss,
	}}
	restoreReader := setSuggestionsReaderForTest(reader)
	t.Cleanup(restoreReader)
	restoreLimiter := setSuggestionsLimiterForTest(&fakeSuggestionsLimiter{allowed: true})
	t.Cleanup(restoreLimiter)

	app := fiber.New()
	app.Get("/search/suggestions", SearchApi.GetSearchSuggestions)
	resp, err := app.Test(httptest.NewRequest("GET", "/search/suggestions?"+params, nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var raw map[string]json.RawMessage
	var body struct {
		Code int                              `json:"code"`
		Data models.SearchSuggestionsResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &fields); err != nil {
		t.Fatal(err)
	}
	if _, found := fields["engine"]; found {
		t.Fatal("engine still exposed")
	}
	encoded, _ := json.Marshal(raw)
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if body.Data.SchemaVersion != 2 {
		t.Fatal("schema must be v2")
	}
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	if body.Data.State != models.SearchSuggestionsStateReady || len(body.Data.Suggestions) != 1 {
		t.Fatalf("unexpected data: %#v", body.Data)
	}
	if reader.lastQuery != "furry" {
		t.Fatalf("reader query = %q", reader.lastQuery)
	}
}

func TestGetSearchSuggestionsRateLimited(t *testing.T) {
	restoreReader := setSuggestionsReaderForTest(&fakeSuggestionsReader{})
	t.Cleanup(restoreReader)
	restoreLimiter := setSuggestionsLimiterForTest(&fakeSuggestionsLimiter{allowed: false, retryAfter: 600})
	t.Cleanup(restoreLimiter)

	app := fiber.New()
	app.Get("/search/suggestions", SearchApi.GetSearchSuggestions)
	resp, err := app.Test(httptest.NewRequest("GET", "/search/suggestions?engine=bing&q=furry", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Retry-After"); got != "600" {
		t.Fatalf("Retry-After = %q", got)
	}
}

type fakeSuggestionsReader struct {
	response  models.SearchSuggestionsResponse
	lastQuery string
}

func (reader *fakeSuggestionsReader) GetSearchSuggestions(_ context.Context, query string) models.SearchSuggestionsResponse {
	reader.lastQuery = query
	return reader.response
}

type fakeSuggestionsLimiter struct {
	allowed    bool
	retryAfter int64
	lastIP     string
}

func (limiter *fakeSuggestionsLimiter) Allow(ip string) (bool, int64) {
	limiter.lastIP = ip
	return limiter.allowed, limiter.retryAfter
}

func TestSuggestionLimiterUsesTrustedClientIP(t *testing.T) {
	cfg := env.GetServerConfig()
	previous := cfg.Server.TrustedProxyCIDRs
	t.Cleanup(func() { cfg.Server.TrustedProxyCIDRs = previous })
	for _, trusted := range []bool{false, true} {
		limiter := &fakeSuggestionsLimiter{allowed: true}
		api := New(&fakeSuggestionsReader{}, limiter)
		app := fiber.New(fiber.Config{TrustProxy: false})
		var peer string
		app.Use(func(c fiber.Ctx) error {
			peer = c.IP()
			cfg.Server.TrustedProxyCIDRs = ""
			if trusted {
				cfg.Server.TrustedProxyCIDRs = peer
			}
			return c.Next()
		})
		app.Get("/search/suggestions", api.GetSearchSuggestions)
		request := httptest.NewRequest(http.MethodGet, "/search/suggestions?q=furry", nil)
		request.Header.Set("X-Forwarded-For", "203.0.113.123")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		want := peer
		if trusted {
			want = "203.0.113.123"
		}
		if limiter.lastIP != want {
			t.Fatalf("trusted=%v: limiter IP=%q want=%q", trusted, limiter.lastIP, want)
		}
	}
}
