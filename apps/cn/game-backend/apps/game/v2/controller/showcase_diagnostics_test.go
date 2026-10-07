package controller_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	controller "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	"github.com/gofurry/gofurry-game-backend/apps/game/v2/showcase"
)

type diagnosticReader struct {
	calls        int
	lang, region string
}

func (r *diagnosticReader) Read(_ context.Context, _ time.Time, lang, region string) (showcase.Inputs, error) {
	r.calls++
	r.lang, r.region = lang, region
	return showcase.Inputs{Diagnostics: map[string][]showcase.Diagnostic{"trending": {
		{GameID: "1", Name: "Other", ExcludedReasons: []string{"adult"}},
		{GameID: "200", Name: "Target", ExcludedReasons: []string{"not_approved"}},
	}}}, nil
}

func TestShowcaseDiagnosticHTTPQuery(t *testing.T) {
	r := &diagnosticReader{}
	api := controller.New(nil, nil, nil, nil).WithShowcase(showcase.New(r, nil), nil, showcase.Signer{})
	app := fiber.New()
	app.Get("/candidates", api.GetShowcaseCandidates)
	res, err := app.Test(httptest.NewRequest("GET", "/candidates?pool=trending&lang=en&region=CN&status=pending_approval&keyword=target&sort=pool&page_size=1", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Data showcase.DiagnosticPage `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || body.Data.Total != 1 || len(body.Data.Items) != 1 || body.Data.Items[0].GameID != "200" || body.Data.Counts.PendingApproval != 1 || r.lang != "en" || r.region != "CN" {
		t.Fatal(res.StatusCode, body)
	}
	for _, query := range []string{"pool=trending&status=typo", "pool=trending&lang=fr", "pool=trending&sort=random", "pool=trending&page_size=101"} {
		res, err := app.Test(httptest.NewRequest("GET", "/candidates?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 400 {
			t.Fatal(query, res.StatusCode)
		}
	}
	if r.calls != 1 {
		t.Fatal("invalid queries read the database", r.calls)
	}
}
