package dao_test

import (
	"context"
	"errors"
	updatesservice "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
)

// Uses the existing disposable migrated DB and rolls back its focused fixture.
func testReleaseNoteVisibilityAndNeighbors(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `DELETE FROM gfn_nav_update_notice;
 INSERT INTO gfn_nav_update_notice (id,title,title_en,body,body_en,summary,summary_en,publication_state,published_at,deleted) VALUES
 (1,'最早','','正文','','','Summary','published','2020-01-01',false),
 (2,'较早','Older','正文','Body','','','published','2020-01-02',false),
 (3,'同日','Same date','正文','Body','','','published','2020-01-02',false),
 (4,'最新','Newest','正文','Body','','','published','2020-01-03',false),
 (5,'草稿','','','','','','draft','2020-01-02',false),
 (6,'定时','','','','','','published',(CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Shanghai') + interval '1 day',false),
 (7,'已删除','','','','','','published','2020-01-02',true),
 (8,'无时间','','','','','','draft',NULL,false),
 (9,'不可见的无时间','','','','','','published',NULL,false);`)
	if err != nil {
		t.Fatal(err)
	}
	svc := updatesservice.New(navsqlc.New(tx))
	index := svc.GetUpdates("en")
	if index.State != "ready" || len(index.Items) != 4 {
		t.Fatalf("public index: %+v", index)
	}
	for i, id := range []int64{4, 3, 2, 1} {
		if index.Items[i].ID != id {
			t.Fatalf("order: %+v", index.Items)
		}
	}
	if index.Items[3].Title != "最早" || index.Items[3].Summary != "Summary" {
		t.Fatal("index locale metadata")
	}
	oldest, err := svc.GetUpdateDetail(1, "en")
	if err != nil || oldest.Item.Body != "正文" {
		t.Fatalf("detail localized body: %+v, %v", oldest, err)
	}
	for _, id := range []int64{5, 6, 7, 8, 9, 404} {
		if _, err := svc.GetUpdateDetail(id, "zh"); !errors.Is(err, updatesservice.ErrNotFound) {
			t.Fatalf("hidden ID %d: %v", id, err)
		}
	}
	for _, tc := range []struct{ id, previous, next int64 }{{1, 0, 2}, {2, 1, 3}, {3, 2, 4}, {4, 3, 0}} {
		detail, err := svc.GetUpdateDetail(tc.id, "en")
		if err != nil {
			t.Fatal(err)
		}
		var previous, next int64
		if detail.Previous != nil {
			previous = detail.Previous.ID
		}
		if detail.Next != nil {
			next = detail.Next.ID
		}
		if previous != tc.previous || next != tc.next {
			t.Fatalf("neighbors %d: %d/%d", tc.id, previous, next)
		}
	}
	// SQL compares China-site wall timestamps, even if a caller changes session timezone.
	if _, err = tx.Exec(ctx, `SET LOCAL TIME ZONE 'America/New_York'`); err != nil {
		t.Fatal(err)
	}
	if got := svc.GetUpdates("zh"); len(got.Items) != 4 {
		t.Fatal("visibility depends on session timezone")
	}
}
