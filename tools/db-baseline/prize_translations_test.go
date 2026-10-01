package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

func TestPrizeTranslationsMigrationPreservesLegacy(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	name := temporaryDatabaseName("gfg", "prize_translations")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs("../../db/game/migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err = goose.UpToContext(ctx, db, dir, 20260916020000); err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO gfg_prize(id,title,"desc",prize,"key",start_time,end_time,create_time,status) VALUES (1,'活动','描述','{"title":"奖品","platform":"Steam","keys":["TEST-A","TEST-B"]}','join','2026-10-01','2026-10-07','2026-10-01',true)`)
	if err != nil {
		t.Fatal(err)
	}
	read := func() string {
		t.Helper()
		var row string
		if err := db.QueryRowContext(ctx, `SELECT (to_jsonb(p)-'title_en'-'desc_en')::text FROM gfg_prize p WHERE id=1`).Scan(&row); err != nil {
			t.Fatal(err)
		}
		return row
	}
	before := read()
	if err = goose.UpToContext(ctx, db, dir, 20261001010000); err != nil {
		t.Fatal(err)
	}
	if read() != before {
		t.Fatal("migration changed existing prize data")
	}
	var defaults bool
	if err = db.QueryRowContext(ctx, `SELECT title_en='' AND desc_en='' FROM gfg_prize WHERE id=1`).Scan(&defaults); err != nil || !defaults {
		t.Fatalf("legacy defaults: %v %v", defaults, err)
	}
	if _, err = db.ExecContext(ctx, `UPDATE gfg_prize SET title_en='Event',desc_en='Description' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if err = goose.DownContext(ctx, db, dir); err == nil || !strings.Contains(err.Error(), "refusing lossy rollback") {
		t.Fatalf("unsafe Down: %v", err)
	}
	if _, err = db.ExecContext(ctx, `UPDATE gfg_prize SET title_en='',desc_en='' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if err = goose.DownContext(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
	if read() != before {
		t.Fatal("safe Down changed original data")
	}
}
