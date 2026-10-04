package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
)

func TestShowcaseProxyForwardsDiagnosticQueryOnly(t *testing.T) {
	requests := make(chan *http.Request, 2)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"code":1,"data":{"total":0,"items":[],"counts":{"all":0,"eligible":0,"pending_approval":0,"blocked":0}}}`)
	}))
	defer upstream.Close()
	cfg := env.BackendServiceConfig{BaseURL: upstream.URL, AdminToken: "test-service-token"}
	app := fiber.New()
	app.Get("/candidates", func(c fiber.Ctx) error { return proxyShowcase(c, cfg) })
	app.Get("/composition", func(c fiber.Ctx) error { return proxyShowcase(c, cfg) })
	for _, operation := range []string{"candidates", "composition"} {
		res, err := app.Test(httptest.NewRequest("GET", "/"+operation+"?pool=trending&lang=en&region=CN&page_num=2&page_size=20&status=pending_approval&keyword=%E7%8B%BC%26%E7%8C%AB&excluded_reason=not_approved&sort=pool&untrusted=ignored", nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}
		req := <-requests
		if req.URL.Path != "/api/v2/game/internal/showcase/"+operation || req.Header.Get("X-GoFurry-Admin-Token") != "test-service-token" || req.URL.Query().Get("untrusted") != "" {
			t.Fatal("proxy contract")
		}
		for key, want := range map[string]string{"status": "pending_approval", "keyword": "狼&猫", "excluded_reason": "not_approved", "sort": "pool"} {
			if operation == "composition" {
				want = ""
			}
			if got := req.URL.Query().Get(key); got != want {
				t.Fatal(operation, key, got, want)
			}
		}
	}
}
