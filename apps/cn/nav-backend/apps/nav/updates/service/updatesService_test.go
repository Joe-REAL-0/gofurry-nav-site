package service

import (
	"errors"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	"github.com/jackc/pgx/v5"
)

func TestGetUpdatesReturnsReadyItems(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2026, 6, 1, 10, 30, 0, 0, time.UTC)
	svc := newUpdatesService(&fakeUpdateNoticeStore{
		items: []models.UpdateNotice{{
			ID:          7,
			Title:       "新时间线",
			TitleEn:     "Timeline refresh",
			Body:        "更新公告改为轻量正文",
			BodyEn:      "The updates page now uses plain text notices.",
			PublishedAt: publishedAt,
			CreateTime:  publishedAt,
			UpdateTime:  publishedAt,
		}},
	}, func() time.Time { return now })

	response := svc.GetUpdates("zh")

	if response.SchemaVersion != models.UpdatesSchemaVersion || !response.GeneratedAt.Equal(now) {
		t.Fatalf("unexpected response metadata: %#v", response)
	}
	if response.State != models.UpdatesStateReady || len(response.Items) != 1 {
		t.Fatalf("expected ready item, got %#v", response)
	}
	if response.Items[0].Title != "新时间线" || response.Items[0].Body == "" {
		t.Fatalf("unexpected item: %#v", response.Items[0])
	}
}

func TestGetUpdatesReturnsEmptyState(t *testing.T) {
	response := newUpdatesService(&fakeUpdateNoticeStore{}, time.Now).GetUpdates("zh")

	if response.State != models.UpdatesStateEmpty || len(response.Items) != 0 {
		t.Fatalf("expected empty response, got %#v", response)
	}
}

func TestGetUpdatesReturnsErrorState(t *testing.T) {
	response := newUpdatesService(&fakeUpdateNoticeStore{err: errors.New("db unavailable")}, time.Now).GetUpdates("zh")

	if response.State != models.UpdatesStateError {
		t.Fatalf("expected error state, got %#v", response)
	}
	if len(response.ReasonMessages) != 1 || response.ReasonMessages[0] != "db unavailable" {
		t.Fatalf("unexpected reason messages: %#v", response.ReasonMessages)
	}
	if len(response.Items) != 0 {
		t.Fatalf("error response should not include items: %#v", response.Items)
	}
}

func TestGetUpdatesReturnsEnglishCopy(t *testing.T) {
	svc := newUpdatesService(&fakeUpdateNoticeStore{
		items: []models.UpdateNotice{{
			ID:          8,
			Title:       "中文标题",
			TitleEn:     "English title",
			Body:        "中文正文",
			BodyEn:      "English body",
			PublishedAt: time.Now(),
		}},
	}, time.Now)

	response := svc.GetUpdates("en")
	if response.Items[0].Title != "English title" || response.Items[0].Body != "English body" {
		t.Fatalf("unexpected localized item: %#v", response.Items[0])
	}
}

type fakeUpdateNoticeStore struct {
	items []models.UpdateNotice
	err   error
}

func (store *fakeUpdateNoticeStore) ListUpdateNotices(limit int) ([]models.UpdateNotice, error) {
	if store.err != nil {
		return nil, store.err
	}
	return append([]models.UpdateNotice(nil), store.items...), nil
}

func (store *fakeUpdateNoticeStore) GetUpdateNotice(id int64) (models.UpdateNotice, *models.UpdateNotice, *models.UpdateNotice, error) {
	if store.err != nil {
		return models.UpdateNotice{}, nil, nil, store.err
	}
	for _, item := range store.items {
		if item.ID == id {
			return item, nil, nil, nil
		}
	}
	return models.UpdateNotice{}, nil, nil, pgx.ErrNoRows
}
func TestReleaseNoteIndependentLocaleFallbackAndDetailProjection(t *testing.T) {
	version, sha := "September notes", "abcdef0"
	notice := models.UpdateNotice{ID: 1, Title: "标题", SummaryEn: "Summary", Body: "正文", Version: &version, CommitSHA: &sha}
	svc := newUpdatesService(&fakeUpdateNoticeStore{items: []models.UpdateNotice{notice}}, time.Now)
	for _, lang := range []string{"zh", " EN "} {
		index := svc.GetUpdates(lang)
		detail, err := svc.GetUpdateDetail(1, lang)
		if err != nil {
			t.Fatal(err)
		}
		item := detail.Item
		if item.Title != "标题" || item.Summary != "Summary" || item.Body != "正文" || *item.Version != version || *item.CommitSHA != sha {
			t.Fatalf("independent fallback failed: %+v", item)
		}
		if index.Items[0].Body != item.Body || index.Items[0].Summary != item.Summary {
			t.Fatal("P1 index body compatibility lost")
		}
		if detail.Previous != nil || detail.Next != nil {
			t.Fatal("empty neighbors must be null")
		}
	}
	if _, err := svc.GetUpdateDetail(404, "en"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if _, err := newUpdatesService(&fakeUpdateNoticeStore{err: errors.New("offline")}, time.Now).GetUpdateDetail(1, "zh"); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("failure masked as missing: %v", err)
	}
}
