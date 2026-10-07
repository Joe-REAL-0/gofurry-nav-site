package controller

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetUpdatesReturnsV2EnvelopeWithoutLegacyURL(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	reader := &fakeUpdatesReader{response: models.UpdatesResponse{
		SchemaVersion: models.UpdatesSchemaVersion,
		GeneratedAt:   now,
		State:         models.UpdatesStateReady,
		Items: []models.UpdateNoticeItem{{
			ID:          1,
			Title:       "公告重构",
			Summary:     "发布摘要",
			PublishedAt: now,
		}},
	}}
	restore := setUpdatesReaderForTest(reader)
	t.Cleanup(restore)

	app := fiber.New()
	app.Get("/updates", UpdatesApi.GetUpdates)
	resp, err := app.Test(httptest.NewRequest("GET", "/updates", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	var body struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if body.Code != common.RETURN_SUCCESS {
		t.Fatalf("unexpected code: %d", body.Code)
	}
	raw := string(body.Data)
	if !strings.Contains(raw, `"state":"ready"`) || !strings.Contains(raw, `"summary":"发布摘要"`) {
		t.Fatalf("unexpected data: %s", raw)
	}
	if strings.Contains(raw, `"url"`) {
		t.Fatalf("legacy url leaked into updates response: %s", raw)
	}
	var projection struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body.Data, &projection); err != nil {
		t.Fatal(err)
	}
	if len(projection.Items) != 1 || len(projection.Items[0]) != 6 {
		t.Fatalf("index projection: %s", raw)
	}
	for _, field := range []string{"id", "title", "summary", "version", "commit_sha", "published_at"} {
		if _, ok := projection.Items[0][field]; !ok {
			t.Fatalf("missing index field %s", field)
		}
	}
	if reader.lastLang != "zh" {
		t.Fatalf("expected default zh lang, got %q", reader.lastLang)
	}
}

type fakeUpdatesReader struct {
	response models.UpdatesResponse
	lastLang string
	lastPage models.UpdatePage
}

func TestPublicIndexPaginationValidation(t *testing.T) {
	for _, tc := range []struct {
		query              string
		status, page, size int
	}{
		{"", 200, 1, 100}, {"?page=2&page_size=21&lang=en", 200, 2, 21},
		{"?page=0", 400, 0, 0}, {"?page=-1", 400, 0, 0}, {"?page=x", 400, 0, 0},
		{"?page_size=101", 400, 0, 0}, {"?page_size=0", 400, 0, 0}, {"?page_size=x", 400, 0, 0},
		{"?page=2147483647&page_size=100", 400, 0, 0},
	} {
		t.Run(tc.query, func(t *testing.T) {
			reader := &fakeUpdatesReader{}
			app := fiber.New()
			app.Get("/updates", New(reader).GetUpdates)
			resp, err := app.Test(httptest.NewRequest("GET", "/updates"+tc.query, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status || reader.lastPage != (models.UpdatePage{Page: tc.page, PageSize: tc.size}) {
				t.Fatalf("status=%d page=%+v", resp.StatusCode, reader.lastPage)
			}
		})
	}
}

func (reader *fakeUpdatesReader) GetUpdates(lang string, pages ...models.UpdatePage) models.UpdatesResponse {
	reader.lastLang = lang
	if len(pages) > 0 {
		reader.lastPage = pages[0]
	}
	return reader.response
}

func (reader *fakeUpdatesReader) GetUpdateDetail(id int64, lang string) (models.UpdateDetailResponse, error) {
	reader.lastLang = lang
	if id == 404 {
		return models.UpdateDetailResponse{}, service.ErrNotFound
	}
	if id == 503 {
		return models.UpdateDetailResponse{}, errors.New("private database error")
	}
	return models.UpdateDetailResponse{SchemaVersion: 1, State: "ready", Item: models.ReleaseNote{ID: id, Title: "Release", Body: "Markdown source"}}, nil
}
func TestReleaseDetailStatusAndPublicFields(t *testing.T) {
	reader := &fakeUpdatesReader{}
	app := fiber.New()
	app.Get("/updates/:id", New(reader).GetUpdateDetail)
	for _, tc := range []struct {
		path   string
		status int
	}{{"1?lang=en", 200}, {"404", 404}, {"-1", 404}, {"bad", 404}, {"503", 503}} {
		resp, err := app.Test(httptest.NewRequest("GET", "/updates/"+tc.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != tc.status {
			t.Fatalf("%s: status=%d", tc.path, resp.StatusCode)
		}
		encoded, _ := json.Marshal(body)
		for _, field := range []string{"publication_state", "deleted", "create_time", "update_time", "private database error"} {
			if strings.Contains(string(encoded), field) {
				t.Fatalf("leaked %s", field)
			}
		}
		if tc.status == 200 {
			data := body["data"].(map[string]any)
			if data["previous"] != nil || data["next"] != nil || reader.lastLang != "en" {
				t.Fatal("detail envelope/locale")
			}
		}
	}
}
