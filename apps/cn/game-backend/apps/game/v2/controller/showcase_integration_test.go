package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	controller "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	"github.com/gofurry/gofurry-game-backend/apps/game/v2/showcase"
	prize "github.com/gofurry/gofurry-game-backend/apps/prize/controller"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/gofurry/gofurry-game-backend/routers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Invoked by Admin's disposable three-database test after a real HTTP publication.
// No production process, collector or cloud client is started by this bridge.
func TestShowcaseCrossServiceHTTPClosure(t *testing.T) {
	dsn := os.Getenv("GOFURRY_SHOWCASE_E2E_GFG_URL")
	id := os.Getenv("GOFURRY_SHOWCASE_E2E_CAMPAIGN_ID")
	addr := os.Getenv("GOFURRY_SHOWCASE_REDIS_ADDR")
	if dsn == "" || id == "" || addr == "" {
		t.Skip("run through TestAdminShowcaseThreeDatabase with disposable Redis")
	}
	ctx := context.Background()
	pool, e := pgxpool.New(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	client := redis.NewClient(&redis.Options{Addr: addr, DB: 13})
	defer client.Close()
	signer := showcase.Signer{TrackingSecret: strings.Repeat("t", 32), HashSecret: strings.Repeat("h", 32)}
	service := showcase.New(showcase.SQLReader{Pool: pool}, client)
	analytics := showcase.NewAnalytics(client, gamesqlc.New(pool), signer, []string{"https://example.test"})
	api := controller.New(nil, nil, nil, nil).WithShowcase(service, analytics, signer)
	cfg := env.GetServerConfig()
	oldAdmin, oldMiddleware := cfg.Admin, cfg.Middleware
	defer func() { cfg.Admin = oldAdmin; cfg.Middleware = oldMiddleware }()
	cfg.Admin.Token = "showcase-integration-service-token"
	cfg.Middleware.Cors.AllowOrigins = "https://example.test"
	cfg.Middleware.Limiter.IsOn = false
	app := routers.NewRouter().Init(pool, api, prize.New(nil))
	call := func(method, path string, body []byte, token bool, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://example.test")
		req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120.0 Safari/537.36")
		if token {
			req.Header.Set("X-GoFurry-Admin-Token", cfg.Admin.Token)
		}
		response, e := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if e != nil {
			t.Fatal(e)
		}
		defer response.Body.Close()
		data, _ := io.ReadAll(response.Body)
		if response.StatusCode != status {
			t.Fatalf("%s: status=%d want=%d %s", path, response.StatusCode, status, data)
		}
		return data
	}
	var response struct {
		Data showcase.Snapshot `json:"data"`
	}
	if e := json.Unmarshal(call("GET", "/api/v2/game/home/showcase?lang=zh&region=CN", nil, false, 200), &response); e != nil {
		t.Fatal(e)
	}
	snap := response.Data
	if len(snap.Items) != 1 || snap.Items[0].CampaignID != id || snap.Items[0].TrackingToken == "" {
		t.Fatal("published Admin Campaign absent", snap)
	}
	if snap.ValidUntil.Sub(snap.GeneratedAt) > 5*time.Minute {
		t.Fatal("TTL")
	}
	if e := json.Unmarshal(call("GET", "/api/v2/game/home/showcase?lang=en&region=CN", nil, false, 200), &response); e != nil || len(response.Data.Items) != 0 || response.Data.Items == nil {
		t.Fatal("locale fallback", e)
	}
	call("GET", "/api/v2/game/internal/showcase/composition", nil, false, 401)
	call("GET", "/api/v2/game/internal/showcase/composition", nil, true, 200)
	call("GET", "/api/v2/game/internal/showcase/candidates?pool=trending", nil, true, 200)
	call("POST", "/api/v2/game/home/showcase/events", []byte(`{"unexpected":true}`), false, 400)
	event := showcase.Event{TrackingToken: snap.Items[0].TrackingToken, SessionID: "11111111-1111-4111-8111-111111111111", Event: "click", Source: "primary"}
	body, _ := json.Marshal(event)
	call("POST", "/api/v2/game/home/showcase/events", body, false, 204)
	call("POST", "/api/v2/game/home/showcase/events", body, false, 204)
	event.Event, event.Source = "impression", ""
	body, _ = json.Marshal(event)
	call("POST", "/api/v2/game/home/showcase/events", body, false, 204)
	for i := 0; i < 3; i++ {
		if e := analytics.Aggregate(ctx); e != nil {
			t.Fatal(e)
		}
	}
	var impressions, clicks, sessions int64
	if e := pool.QueryRow(ctx, `SELECT valid_impressions,qualified_clicks,session_estimate FROM gfg_showcase_daily_stat WHERE subject_key=$1`, "campaign:"+id).Scan(&impressions, &clicks, &sessions); e != nil || impressions != 1 || clicks != 1 || sessions != 1 {
		t.Fatal("absolute aggregate", impressions, clicks, sessions, e)
	}
	var raw string
	if e := pool.QueryRow(ctx, `SELECT row_to_json(s)::text FROM gfg_showcase_daily_stat s WHERE subject_key=$1`, "campaign:"+id).Scan(&raw); e != nil {
		t.Fatal(e)
	}
	for _, private := range []string{event.SessionID, "Mozilla", "127.0.0.1", event.TrackingToken} {
		if strings.Contains(raw, private) {
			t.Fatal("private event material persisted")
		}
	}
	client.Close()
	call("POST", "/api/v2/game/home/showcase/events", body, false, 204)
}
