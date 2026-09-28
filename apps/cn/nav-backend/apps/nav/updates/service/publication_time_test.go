package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestPublicReleasePublicationTimePreservesChinaWallClock(t *testing.T) {
	// pgx decodes a timestamp WITHOUT time zone into UTC-shaped calendar fields.
	// The Release Notes domain defines those fields as Asia/Shanghai wall time.
	wall := pgtype.Timestamp{Time: time.Date(2026, 9, 1, 0, 30, 0, 0, time.UTC), Valid: true}
	notice := noticeModel(navsqlc.GfnNavUpdateNotice{ID: 1, Title: "Release", PublishedAt: wall})
	svc := newUpdatesService(&fakeUpdateNoticeStore{items: []models.UpdateNotice{notice}}, time.Now)
	detail, err := svc.GetUpdateDetail(1, "zh")
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]time.Time{
		"index":    svc.GetUpdates("zh").Items[0].PublishedAt,
		"detail":   detail.Item.PublishedAt,
		"neighbor": localizeNeighbor(&notice, "zh").PublishedAt,
	} {
		encoded, err := json.Marshal(value)
		if err != nil || string(encoded) != `"2026-09-01T00:30:00+08:00"` {
			t.Errorf("%s incorrectly labels the stored China wall clock: %s (%v)", name, encoded, err)
		}
		if !value.Equal(time.Date(2026, 8, 31, 16, 30, 0, 0, time.UTC)) {
			t.Errorf("%s does not represent the actual publication instant", name)
		}
	}
}
