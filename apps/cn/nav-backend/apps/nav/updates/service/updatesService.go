package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
)

const defaultUpdatesLimit = 100

type updateNoticeStore interface {
	ListUpdateNotices(limit, offset int) ([]models.UpdateNotice, int64, error)
	GetUpdateNotice(id int64) (models.UpdateNotice, *models.UpdateNotice, *models.UpdateNotice, error)
}

type sqlcUpdateNoticeStore struct {
	queries *navsqlc.Queries
}

func newSQLCUpdateNoticeStore(queries *navsqlc.Queries) *sqlcUpdateNoticeStore {
	return &sqlcUpdateNoticeStore{queries: queries}
}

func (store *sqlcUpdateNoticeStore) ListUpdateNotices(limit, offset int) ([]models.UpdateNotice, int64, error) {
	if limit <= 0 {
		limit = defaultUpdatesLimit
	}
	asOf, err := store.queries.UpdateNoticeClock(context.Background())
	if err != nil {
		return nil, 0, err
	}
	total, err := store.queries.CountPublicUpdateNotices(context.Background(), asOf)
	if err != nil {
		return nil, 0, err
	}
	rows, err := store.queries.ListPublicUpdateNotices(context.Background(), navsqlc.ListPublicUpdateNoticesParams{AsOf: asOf, RowLimit: int32(limit), RowOffset: int32(offset)})
	if err != nil {
		return nil, 0, err
	}
	notices := make([]models.UpdateNotice, 0, len(rows))
	for _, row := range rows {
		notices = append(notices, noticeModel(row))
	}
	return notices, total, nil
}

type updatesService struct {
	store updateNoticeStore
	now   func() time.Time
}

var (
	updatesSingleton = &updatesService{}
	updatesMu        sync.Mutex
)

func GetUpdatesService() *updatesService {
	updatesMu.Lock()
	defer updatesMu.Unlock()
	if updatesSingleton.now == nil {
		updatesSingleton.now = time.Now
	}
	return updatesSingleton
}

func newUpdatesService(store updateNoticeStore, now func() time.Time) *updatesService {
	return &updatesService{store: store, now: now}
}

func New(queries *navsqlc.Queries) *updatesService {
	return newUpdatesService(newSQLCUpdateNoticeStore(queries), time.Now)
}

func (svc *updatesService) GetUpdates(lang string, pages ...models.UpdatePage) models.UpdatesResponse {
	lang = normalizeLang(lang)
	page := models.UpdatePage{Page: 1, PageSize: defaultUpdatesLimit}
	if len(pages) > 0 {
		page = pages[0]
	}
	if page.Page <= 0 {
		page.Page = 1
	}
	if page.PageSize <= 0 || page.PageSize > defaultUpdatesLimit {
		page.PageSize = defaultUpdatesLimit
	}
	response := models.UpdatesResponse{
		Page: page.Page, PageSize: page.PageSize,
		SchemaVersion: models.UpdatesSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.UpdatesStateEmpty,
		Items:         []models.UpdateNoticeItem{},
	}

	offset := (page.Page - 1) * page.PageSize
	notices, total, err := svc.source().ListUpdateNotices(page.PageSize, offset)
	if err != nil {
		response.State = models.UpdatesStateError
		response.ReasonMessages = []string{err.Error()}
		return response
	}
	response.Total = total
	response.HasMore = int64(offset+len(notices)) < total && len(notices) > 0

	if len(notices) == 0 {
		return response
	}

	response.State = models.UpdatesStateReady
	response.Items = make([]models.UpdateNoticeItem, 0, len(notices))
	for _, notice := range notices {
		title, _ := localizeNotice(notice, lang)
		response.Items = append(response.Items, models.UpdateNoticeItem{
			Summary: localizeSummary(notice, lang), Version: notice.Version, CommitSHA: notice.CommitSHA,
			ID:          notice.ID,
			Title:       title,
			PublishedAt: notice.PublishedAt,
		})
	}
	return response
}

func (svc *updatesService) source() updateNoticeStore {
	return svc.store
}

func (svc *updatesService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func normalizeLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		return "en"
	}
	return "zh"
}

func localizeNotice(notice models.UpdateNotice, lang string) (string, string) {
	if lang == "en" {
		return firstNonEmpty(notice.TitleEn, notice.Title), firstNonEmpty(notice.BodyEn, notice.Body)
	}
	return firstNonEmpty(notice.Title, notice.TitleEn), firstNonEmpty(notice.Body, notice.BodyEn)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

var ErrNotFound = errors.New("release note not found")

func noticeModel(row navsqlc.GfnNavUpdateNotice) models.UpdateNotice {
	return models.UpdateNotice{ID: row.ID, Title: row.Title, TitleEn: row.TitleEn, Body: row.Body, BodyEn: row.BodyEn,
		Summary: row.Summary, SummaryEn: row.SummaryEn, Version: row.Version, CommitSHA: row.CommitSha,
		PublishedAt: releasePublicationTime(row.PublishedAt.Time), CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time, Deleted: row.Deleted}
}

// Release dates are contemporary China-site wall timestamps, not UTC instants.
// Attach their UTC+8 offset before JSON serialization; never shift stored fields.
var releasePublicationZone = time.FixedZone("Asia/Shanghai", 8*60*60)

func releasePublicationTime(wall time.Time) time.Time {
	if wall.IsZero() {
		return wall
	}
	return time.Date(wall.Year(), wall.Month(), wall.Day(), wall.Hour(), wall.Minute(), wall.Second(), wall.Nanosecond(), releasePublicationZone)
}

func (store *sqlcUpdateNoticeStore) GetUpdateNotice(id int64) (models.UpdateNotice, *models.UpdateNotice, *models.UpdateNotice, error) {
	ctx := context.Background()
	asOf, err := store.queries.UpdateNoticeClock(ctx)
	if err != nil {
		return models.UpdateNotice{}, nil, nil, err
	}
	row, err := store.queries.GetPublicUpdateNotice(ctx, navsqlc.GetPublicUpdateNoticeParams{ID: id, AsOf: asOf})
	if err != nil {
		return models.UpdateNotice{}, nil, nil, err
	}
	// Both bounded neighbors use the same database clock as the selected release.
	previous, err := store.queries.GetPreviousPublicUpdateNotice(ctx, navsqlc.GetPreviousPublicUpdateNoticeParams{ID: id, PublishedAt: row.PublishedAt, AsOf: asOf})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return models.UpdateNotice{}, nil, nil, err
	}
	var older, newer *models.UpdateNotice
	if err == nil {
		notice := noticeModel(previous)
		older = &notice
	}
	next, err := store.queries.GetNextPublicUpdateNotice(ctx, navsqlc.GetNextPublicUpdateNoticeParams{ID: id, PublishedAt: row.PublishedAt, AsOf: asOf})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return models.UpdateNotice{}, nil, nil, err
	}
	if err == nil {
		notice := noticeModel(next)
		newer = &notice
	}
	return noticeModel(row), older, newer, nil
}

func localizeSummary(notice models.UpdateNotice, lang string) string {
	if lang == "en" {
		return firstNonEmpty(notice.SummaryEn, notice.Summary)
	}
	return firstNonEmpty(notice.Summary, notice.SummaryEn)
}

func (svc *updatesService) GetUpdateDetail(id int64, lang string) (models.UpdateDetailResponse, error) {
	response := models.UpdateDetailResponse{SchemaVersion: models.UpdatesSchemaVersion, GeneratedAt: svc.clock()(), State: models.UpdatesStateReady}
	if id <= 0 {
		return response, ErrNotFound
	}
	notice, previous, next, err := svc.source().GetUpdateNotice(id)
	if errors.Is(err, pgx.ErrNoRows) {
		return response, ErrNotFound
	}
	if err != nil {
		return response, err
	}
	lang = normalizeLang(lang)
	title, body := localizeNotice(notice, lang)
	response.Item = models.ReleaseNote{ID: notice.ID, Title: title, Summary: localizeSummary(notice, lang), Body: body, Version: notice.Version, CommitSHA: notice.CommitSHA, PublishedAt: notice.PublishedAt}
	response.Previous, response.Next = localizeNeighbor(previous, lang), localizeNeighbor(next, lang)
	return response, nil
}

func localizeNeighbor(notice *models.UpdateNotice, lang string) *models.ReleaseNoteNeighbor {
	if notice == nil {
		return nil
	}
	title, _ := localizeNotice(*notice, lang)
	return &models.ReleaseNoteNeighbor{ID: notice.ID, Title: title, Version: notice.Version, PublishedAt: notice.PublishedAt}
}
