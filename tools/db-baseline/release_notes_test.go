package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

func TestReleaseNotesMigrationPreservesLegacyAndGuardsDown(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	name := temporaryDatabaseName("gfn", "release_notes")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs("../../db/nav/migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err = goose.UpToContext(ctx, db, dir, 20260913010000); err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO gfn_nav_update_notice(id,title,title_en,body,body_en,published_at,create_time,update_time,deleted) VALUES
 (7,'中文标题','English title','# Markdown 原文','English body','2026-09-01 12:00:00','2026-08-01','2026-08-02',false),
 (8,'保留删除','Deleted','旧正文','Old body','2099-01-01','2026-08-03','2026-08-04',true)`)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := func() string {
		t.Helper()
		var value string
		err := db.QueryRowContext(ctx, `SELECT jsonb_agg(jsonb_build_array(id,title,title_en,body,body_en,published_at,create_time,update_time,deleted) ORDER BY id)::text FROM gfn_nav_update_notice`).Scan(&value)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	// This test owns the Release Notes Down contract, not later migrations.
	if err = goose.UpToContext(ctx, db, dir, 20260928010000); err != nil {
		t.Fatal(err)
	}
	if snapshot() != before {
		t.Fatal("legacy identity/content/timestamps/deleted changed")
	}
	var count int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM gfn_nav_update_notice WHERE publication_state='published' AND version IS NULL AND commit_sha IS NULL AND summary='' AND summary_en=''`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("backfill: %d %v", count, err)
	}
	for _, statement := range []string{
		`UPDATE gfn_nav_update_notice SET publication_state='scheduled' WHERE id=7`,
		`UPDATE gfn_nav_update_notice SET commit_sha='ABCDEF0' WHERE id=7`,
		`UPDATE gfn_nav_update_notice SET commit_sha='abcdef' WHERE id=7`,
		`UPDATE gfn_nav_update_notice SET version=' ' WHERE id=7`,
	} {
		if _, err = db.ExecContext(ctx, statement); err == nil {
			t.Fatalf("constraint missing: %s", statement)
		}
	}
	_, err = db.ExecContext(ctx, `INSERT INTO gfn_nav_update_notice(id,title,title_en,body,body_en) VALUES (9,'','','','')`)
	if err != nil {
		t.Fatal(err)
	}
	var state string
	var missing bool
	if err = db.QueryRowContext(ctx, `SELECT publication_state,published_at IS NULL FROM gfn_nav_update_notice WHERE id=9`).Scan(&state, &missing); err != nil || state != "draft" || !missing {
		t.Fatal("nullable new draft defaults")
	}
	if err = goose.DownContext(ctx, db, dir); err == nil || !strings.Contains(err.Error(), "published_at is NULL") {
		t.Fatalf("unsafe Down: %v", err)
	}
	if err = db.QueryRowContext(ctx, `SELECT publication_state FROM gfn_nav_update_notice WHERE id=9`).Scan(&state); err != nil || state != "draft" {
		t.Fatal("failed rollback modified schema/data")
	}
	if _, err = db.ExecContext(ctx, `DELETE FROM gfn_nav_update_notice WHERE id=9`); err != nil {
		t.Fatal(err)
	}
	if err = goose.DownContext(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
	if snapshot() != before {
		t.Fatal("safe rollback changed old data")
	}
	if err = goose.UpToContext(ctx, db, dir, 20260928010000); err != nil {
		t.Fatal(err)
	}
	if snapshot() != before {
		t.Fatal("reapply changed old data")
	}
}
