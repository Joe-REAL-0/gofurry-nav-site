package controller

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/recommendations/models"
)

type fakeReader struct {
	id    int64
	lang  string
	limit int
	calls int
}

func (f *fakeReader) GetRecommendations(id int64, lang string, limit int) models.Response {
	f.id = id
	f.lang = lang
	f.limit = limit
	f.calls++
	return models.Response{SchemaVersion: 1, SiteID: id, State: "ready"}
}
func TestRecommendationsRequestContract(t *testing.T) {
	for _, item := range []struct {
		url    string
		status int
		lang   string
		limit  int
	}{
		{"/sites/41/recommendations", 200, "zh", 8},
		{"/sites/41/recommendations?lang=en&limit=3", 200, "en", 3},
		{"/sites/nope/recommendations", 400, "", 0},
		{"/sites/0/recommendations", 400, "", 0},
		{"/sites/41/recommendations?lang=fr", 400, "", 0},
		{"/sites/41/recommendations?limit=banana", 400, "", 0},
	} {
		t.Run(item.url, func(t *testing.T) {
			reader := &fakeReader{}
			app := fiber.New()
			app.Get("/sites/:siteId/recommendations", New(reader).GetRecommendations)
			response, err := app.Test(httptest.NewRequest("GET", item.url, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != item.status {
				t.Fatal(response.StatusCode)
			}
			if item.status == 200 {
				var body struct {
					Data models.Response `json:"data"`
				}
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Data.SiteID != 41 || reader.id != 41 || reader.lang != item.lang || reader.limit != item.limit || reader.calls != 1 {
					t.Fatalf("reader=%+v data=%+v", reader, body)
				}
			} else if reader.calls != 0 {
				t.Fatal("invalid request reached service")
			}
		})
	}
}
