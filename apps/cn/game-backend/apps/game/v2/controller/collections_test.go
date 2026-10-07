package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
)

type publicCollectionReader struct {
	lang          string
	limit, offset int64
	err           error
	calls         int
}

func (r *publicCollectionReader) CountPublishedCollections(context.Context, v2models.CollectionQuery) (int64, error) {
	return 1, r.err
}
func (r *publicCollectionReader) ListPublishedCollections(_ context.Context, query v2models.CollectionQuery) ([]v2models.CollectionRecord, error) {
	r.limit, r.offset = query.PageSize, (query.Page-1)*query.PageSize
	return []v2models.CollectionRecord{{ID: 1, Code: "public"}}, r.err
}
func (r *publicCollectionReader) GetPublishedCollection(_ context.Context, code string) (*v2models.CollectionRecord, error) {
	r.calls++
	if code != "public" {
		return nil, r.err
	}
	return &v2models.CollectionRecord{ID: 1, Code: code}, r.err
}
func (r *publicCollectionReader) ListPublishedCollectionHomeSlots(context.Context, string) ([]v2models.CollectionRecord, error) {
	return []v2models.CollectionRecord{}, r.err
}
func (r *publicCollectionReader) LoadCollectionProjectionGames(_ context.Context, _ []int64, lang string) (v2models.CollectionGames, error) {
	r.lang = lang
	return v2models.CollectionGames{Memberships: []v2models.CollectionMembership{{CollectionID: 1, GameID: 1}}, Games: []v2models.CollectionProjectionGame{{Site: v2models.GameV2SiteRecord{ID: 1}, Adult: true}}}, r.err
}

func TestCollectionHTTPContract(t *testing.T) {
	r := &publicCollectionReader{}
	s := v2service.NewCollectionService(r, nil)
	s.Now = func() time.Time { return time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC) }
	api := New(nil, nil, nil, nil).WithCollections(s)
	app := fiber.New()
	app.Get("/collections/home", api.GetCollectionHome)
	app.Get("/collections", api.GetCollections)
	app.Get("/collections/:code", api.GetCollection)
	for _, tc := range []struct {
		query, lang string
		page, size  int64
		count       int
	}{
		{"", "zh", 1, 24, 0},
		{"?lang=en&mode=nsfw&page=2&page_size=100", "en", 2, 60, 1},
		{"?lang=fr&mode=invalid&page=broken&page_size=-2", "zh", 1, 24, 0},
		{"?page=99999999999999999999999&page_size=99999999999999999999999", "zh", 1, 24, 0},
	} {
		res, e := app.Test(httptest.NewRequest("GET", "/collections"+tc.query, nil))
		if e != nil {
			t.Fatal(e)
		}
		var body struct {
			Code int                      `json:"code"`
			Data v2models.CollectionIndex `json:"data"`
		}
		e = json.NewDecoder(res.Body).Decode(&body)
		res.Body.Close()
		if e != nil || res.StatusCode != 200 || body.Code != 1 || body.Data.SchemaVersion != 1 || body.Data.AsOfDate != "2026-10-06" || body.Data.Page != tc.page || body.Data.PageSize != tc.size || r.lang != tc.lang || r.limit != tc.size || r.offset != (tc.page-1)*tc.size || body.Data.Items[0].VisibleGameCount != tc.count {
			t.Fatalf("%s: %+v err=%v", tc.query, body, e)
		}
	}
	for _, path := range []string{"/collections/home", "/collections/public"} {
		res, e := app.Test(httptest.NewRequest("GET", path, nil))
		if e != nil {
			t.Fatal(e)
		}
		res.Body.Close()
		if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-store" {
			t.Fatal(path, res.StatusCode)
		}
	}
	for _, code := range []string{"draft", "archived", "missing", "bad--code"} {
		res, e := app.Test(httptest.NewRequest("GET", "/collections/"+code, nil))
		if e != nil {
			t.Fatal(e)
		}
		var body struct {
			Code int    `json:"code"`
			Data string `json:"data"`
		}
		e = json.NewDecoder(res.Body).Decode(&body)
		res.Body.Close()
		if e != nil || res.StatusCode != 404 || body.Data != "Collection not found" {
			t.Fatal(code, res.StatusCode, body)
		}
	}
	if r.calls != 4 {
		t.Fatal("invalid code reached DAO or home was treated as detail", r.calls)
	}
	r.err = errors.New("private SQL failure")
	res, e := app.Test(httptest.NewRequest("GET", "/collections", nil))
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	var body struct {
		Data string `json:"data"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	if res.StatusCode != 503 || body.Data != "Collections unavailable" {
		t.Fatal("DB error disclosure", res.StatusCode, body)
	}
}

func TestCollectionInvalidDiscoveryCriteria(t *testing.T) {
	api := New(nil, nil, nil, nil).WithCollections(v2service.NewCollectionService(&publicCollectionReader{}, nil))
	app := fiber.New()
	app.Get("/collections", api.GetCollections)
	for _, query := range []string{"?phase=bad", "?sort=bad", "?phase=all&sort=bad"} {
		res, err := app.Test(httptest.NewRequest("GET", "/collections"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 400 || res.Header.Get("Cache-Control") != "no-store" {
			t.Fatal(query, res.StatusCode)
		}
	}
}
