package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"io"
	"net/http/httptest"
	"testing"
)

func TestGlobalLimiterUsesTrustedClientIP(t *testing.T) {
	cfg := env.GetServerConfig()
	oldServer, oldMiddleware, oldWAF := cfg.Server, cfg.Middleware, cfg.Waf
	defer func() { cfg.Server = oldServer; cfg.Middleware = oldMiddleware; cfg.Waf = oldWAF }()
	cfg.Server.Mode = "release"
	cfg.Middleware.Cors.AllowOrigins = "https://example.test"
	cfg.Middleware.Limiter = env.LimiterConfig{IsOn: true, MaxRequests: 1, Expiration: 60}
	cfg.Waf.WafSwitch = false
	for _, trusted := range []bool{false, true} {
		cfg.Server.TrustedProxyCIDRs = nil
		if trusted {
			cfg.Server.TrustedProxyCIDRs = []string{"0.0.0.0/32"}
		}
		app := fiber.New(fiber.Config{TrustProxy: false})
		registerMiddlewares(app)
		app.Get("/probe", func(c fiber.Ctx) error { return c.SendString(util.GetClientIP(c)) })
		for i, ip := range []string{"192.0.2.1", "192.0.2.2"} {
			req := httptest.NewRequest("GET", "/probe", nil)
			req.Header.Set("X-Forwarded-For", ip)
			response, e := app.Test(req)
			if e != nil {
				t.Fatal(e)
			}
			body, _ := io.ReadAll(response.Body)
			response.Body.Close()
			want := 200
			if !trusted && i == 1 {
				want = 429
			}
			if response.StatusCode != want {
				t.Fatalf("trusted=%v status=%d wanted=%d body=%s", trusted, response.StatusCode, want, body)
			}
			if trusted && string(body) != ip {
				t.Fatal("trusted helper not used", string(body))
			}
		}
	}
}
